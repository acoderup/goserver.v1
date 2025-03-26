package balancequeue

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

type A struct {
	Name string
}

func (a *A) BalanceQueueHandler() {

}

func TestOne(t *testing.T) {
	q := New(5)

	var es []Element
	go func() {
		for {
			q.Update()
			fmt.Println(q)
			time.Sleep(time.Second)
			e := &A{Name: fmt.Sprint(time.Now().Unix())}
			es = append(es, e)
			q.Push(e)
			if rand.Intn(10) > 5 && len(es) >= 2 {
				for _, v := range es[:2] {
					q.Pop(v)
				}
				es = es[2:]
			}
		}
	}()

	time.Sleep(time.Minute)
}
