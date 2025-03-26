package rabbitmq

import (
	"context"

	"github.com/acoderupacoderup/goserver.v1/core/broker"
)

/*
 context参数
*/

type durableQueueKey struct{}
type headersKey struct{}
type queueArgumentsKey struct{}
type prefetchCountKey struct{}
type prefetchGlobalKey struct{}
type exchangeKey struct{}
type durableExchange struct{}
type requeueOnErrorKey struct{}
type deliveryMode struct{}
type priorityKey struct{}
type externalAuth struct{}
type ackSuccessKey struct{}
type subscribeContextKey struct{}

//============================
// broker.SubscribeOption
//============================

// DurableQueue creates a durable queue when subscribing.
func DurableQueue() broker.SubscribeOption {
	return setSubscribeOption(durableQueueKey{}, true)
}

// Headers adds headers used by the headers exchange
func Headers(h map[string]interface{}) broker.SubscribeOption {
	return setSubscribeOption(headersKey{}, h)
}

// QueueArguments sets arguments for queue creation
func QueueArguments(h map[string]interface{}) broker.SubscribeOption {
	return setSubscribeOption(queueArgumentsKey{}, h)
}

// RequeueOnError calls Nack(muliple:false, requeue:true) on amqp delivery when handler returns error
func RequeueOnError() broker.SubscribeOption {
	return setSubscribeOption(requeueOnErrorKey{}, true)
}

// SubscribeContext set the context for broker.SubscribeOption
func SubscribeContext(ctx context.Context) broker.SubscribeOption {
	return setSubscribeOption(subscribeContextKey{}, ctx)
}

// AckOnSuccess will automatically acknowledge messages when no error is returned
func AckOnSuccess() broker.SubscribeOption {
	return setSubscribeOption(ackSuccessKey{}, true)
}

//============================
// broker.Option
//============================

// DurableExchange is an option to set the Exchange to be durable
func DurableExchange() broker.Option {
	return setBrokerOption(durableExchange{}, true)
}

// ExchangeName is an option to set the ExchangeName
func ExchangeName(e string) broker.Option {
	return setBrokerOption(exchangeKey{}, e)
}

// PrefetchCount ...
func PrefetchCount(c int) broker.Option {
	return setBrokerOption(prefetchCountKey{}, c)
}

func ExternalAuth() broker.Option {
	return setBrokerOption(externalAuth{}, ExternalAuthentication{})
}

// PrefetchGlobal creates a durable queue when subscribing.
func PrefetchGlobal() broker.Option {
	return setBrokerOption(prefetchGlobalKey{}, true)
}

//============================
// broker.PublishOption
//============================

// DeliveryMode sets a delivery mode for publishing
func DeliveryMode(value uint8) broker.PublishOption {
	return setPublishOption(deliveryMode{}, value)
}

// Priority sets a priority level for publishing
func Priority(value uint8) broker.PublishOption {
	return setPublishOption(priorityKey{}, value)
}
