// Package rabbitmq provides a RabbitMQ broker
package rabbitmq

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/acoderup/goserver/core/broker"
	"github.com/streadway/amqp"
)

// publication Event
type publication struct {
	d   amqp.Delivery
	m   *broker.Message
	t   string // RouteKey
	err error
}

func (p *publication) Ack() error {
	return p.d.Ack(false)
}

func (p *publication) Error() error {
	return p.err
}

func (p *publication) Topic() string {
	return p.t
}

func (p *publication) Message() *broker.Message {
	return p.m
}

type subscriber struct {
	mtx    sync.Mutex
	mayRun bool

	// config
	opts         broker.SubscribeOptions
	topic        string
	durableQueue bool
	queueArgs    map[string]interface{}
	headers      map[string]interface{}

	ch *rabbitMQChannel
	r  *myBroker
	fn func(msg amqp.Delivery)
}

func (s *subscriber) Options() broker.SubscribeOptions {
	return s.opts
}

func (s *subscriber) Topic() string {
	return s.topic
}

func (s *subscriber) Unsubscribe() error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.mayRun = false
	if s.ch != nil {
		return s.ch.Close()
	}
	return nil
}

func (s *subscriber) resubscribe() {
	minResubscribeDelay := 100 * time.Millisecond
	maxResubscribeDelay := 30 * time.Second
	expFactor := time.Duration(2)
	reSubscribeDelay := minResubscribeDelay
	//loop until unsubscribe
	for {
		s.mtx.Lock()
		mayRun := s.mayRun
		s.mtx.Unlock()
		if !mayRun {
			// we are unsubscribed, showdown routine
			return
		}

		select {
		//check shutdown case
		case <-s.r.conn.close:
			//yep, its shutdown case
			return
			//wait until we reconect to rabbit
		case <-s.r.conn.WaitConnection:
		}

		// it may crash (panic) in case of Consume without connection, so recheck it
		s.r.mtx.Lock()
		if !s.r.conn.connected {
			s.r.mtx.Unlock()
			continue
		}

		ch, sub, err := s.r.conn.Consume(
			s.opts.Queue,
			s.topic,
			s.headers,
			s.queueArgs,
			s.opts.AutoAck,
			s.durableQueue,
		)

		s.r.mtx.Unlock()
		switch err {
		case nil:
			reSubscribeDelay = minResubscribeDelay
			s.mtx.Lock()
			s.ch = ch
			s.mtx.Unlock()
		default:
			if reSubscribeDelay > maxResubscribeDelay {
				reSubscribeDelay = maxResubscribeDelay
			}
			time.Sleep(reSubscribeDelay)
			reSubscribeDelay *= expFactor
			continue
		}
		for d := range sub {
			s.r.wg.Add(1)
			s.fn(d)
			s.r.wg.Done()
		}
	}
}

type myBroker struct {
	conn           *rabbitMQConn
	addrs          []string
	opts           broker.Options
	prefetchCount  int
	prefetchGlobal bool
	mtx            sync.Mutex
	wg             sync.WaitGroup
}

func NewBroker(opts ...broker.Option) broker.Broker {
	options := broker.Options{
		Context: context.Background(),
	}

	for _, o := range opts {
		o(&options)
	}

	return &myBroker{
		addrs: options.Addrs,
		opts:  options,
	}
}

// Publish 发送消息
// 可设置消息的持久化、优先级、过期时间等属性
func (r *myBroker) Publish(topic string, msg *broker.Message, opts ...broker.PublishOption) error {
	m := amqp.Publishing{
		Body:    msg.Body,
		Headers: amqp.Table{},
	}

	options := broker.PublishOptions{}
	for _, o := range opts {
		o(&options)
	}

	if options.Context != nil {
		if value, ok := options.Context.Value(deliveryMode{}).(uint8); ok {
			m.DeliveryMode = value
		}

		if value, ok := options.Context.Value(priorityKey{}).(uint8); ok {
			m.Priority = value
		}
	}

	for k, v := range msg.Header {
		m.Headers[k] = v
	}

	if r.conn == nil {
		return errors.New("connection is nil")
	}

	return r.conn.Publish(r.conn.exchange.Name, topic, m)
}

// Subscribe 创建消费者
// 默认开启自动确认消息, 使用broker.DisableAutoAck()关闭
// 关闭自动确认消息后使用 AckOnSuccess() 可以根据handler的错误信息自动确认或者拒绝消息
func (r *myBroker) Subscribe(topic string, handler broker.Handler, opts ...broker.SubscribeOption) (broker.Subscriber, error) {
	var ackSuccess bool

	if r.conn == nil {
		return nil, errors.New("not connected")
	}

	opt := broker.SubscribeOptions{
		AutoAck: true,
	}

	for _, o := range opts {
		o(&opt)
	}

	// Make sure context is setup
	if opt.Context == nil {
		opt.Context = context.Background()
	}

	ctx := opt.Context
	if subscribeContext, ok := ctx.Value(subscribeContextKey{}).(context.Context); ok && subscribeContext != nil {
		ctx = subscribeContext
	}

	var requeueOnError bool
	requeueOnError, _ = ctx.Value(requeueOnErrorKey{}).(bool)

	var durableQueue bool
	durableQueue, _ = ctx.Value(durableQueueKey{}).(bool)

	var qArgs map[string]interface{}
	if qa, ok := ctx.Value(queueArgumentsKey{}).(map[string]interface{}); ok {
		qArgs = qa
	}

	var headers map[string]interface{}
	if h, ok := ctx.Value(headersKey{}).(map[string]interface{}); ok {
		headers = h
	}

	if su, ok := ctx.Value(ackSuccessKey{}).(bool); ok && su {
		opt.AutoAck = false
		ackSuccess = true
	}

	fn := func(msg amqp.Delivery) {
		header := make(map[string]string)
		for k, v := range msg.Headers {
			header[k], _ = v.(string)
		}
		m := &broker.Message{
			Header: header,
			Body:   msg.Body,
		}
		p := &publication{d: msg, m: m, t: msg.RoutingKey}
		p.err = handler(p)
		// AutoAck为false时根据有没有错误自动确认或者重新入队
		if p.err == nil && ackSuccess && !opt.AutoAck {
			msg.Ack(false)
		} else if p.err != nil && !opt.AutoAck {
			msg.Nack(false, requeueOnError)
		}
	}

	s := &subscriber{topic: topic, opts: opt, mayRun: true, r: r,
		durableQueue: durableQueue, fn: fn, headers: headers, queueArgs: qArgs}

	go s.resubscribe()

	return s, nil
}

func (r *myBroker) Options() broker.Options {
	return r.opts
}

func (r *myBroker) String() string {
	return "rabbitmq"
}

func (r *myBroker) Address() string {
	if len(r.addrs) > 0 {
		return r.addrs[0]
	}
	return ""
}

func (r *myBroker) Init(opts ...broker.Option) error {
	for _, o := range opts {
		o(&r.opts)
	}
	r.addrs = r.opts.Addrs
	return nil
}

// Connect 建立连接
func (r *myBroker) Connect() error {
	if r.conn == nil {
		r.conn = newRabbitMQConn(r.getExchange(), r.opts.Addrs, r.getPrefetchCount(), r.getPrefetchGlobal())
	}

	conf := defaultAmqpConfig

	if auth, ok := r.opts.Context.Value(externalAuth{}).(ExternalAuthentication); ok {
		conf.SASL = []amqp.Authentication{&auth}
	}

	conf.TLSClientConfig = r.opts.TLSConfig

	return r.conn.Connect(r.opts.Secure, &conf)
}

// Disconnect 停止消费者后断开连接
// 等待消费者停止消费
func (r *myBroker) Disconnect() error {
	if r.conn == nil {
		return errors.New("connection is nil")
	}
	ret := r.conn.Close()
	r.wg.Wait() // wait all goroutines
	return ret
}

func (r *myBroker) getExchange() Exchange {

	ex := DefaultExchange

	if e, ok := r.opts.Context.Value(exchangeKey{}).(string); ok {
		ex.Name = e
	}

	if d, ok := r.opts.Context.Value(durableExchange{}).(bool); ok {
		ex.Durable = d
	}

	return ex
}

func (r *myBroker) getPrefetchCount() int {
	if e, ok := r.opts.Context.Value(prefetchCountKey{}).(int); ok {
		return e
	}
	return DefaultPrefetchCount
}

func (r *myBroker) getPrefetchGlobal() bool {
	if e, ok := r.opts.Context.Value(prefetchGlobalKey{}).(bool); ok {
		return e
	}
	return DefaultPrefetchGlobal
}
