package balancequeue

import (
	"fmt"
	"strings"
)

// 平衡队列

type Element interface {
	BalanceQueueHandler()
}

type elementWrapper struct {
	F func()
}

func (e *elementWrapper) BalanceQueueHandler() {
	e.F()
}

func ElementWrapper(f func()) Element {
	return &elementWrapper{F: f}
}

type group struct {
	Array    []Element
	queuePos int
}

type groupArray struct {
	queue []*group
}

type BalanceQueue struct {
	index  int      // 循环索引
	groups []*group // 固定的分组，长度不变，每次Update触发一个分组
	tables []*groupArray
	pool   map[Element]*group
}

// New 创建一个平衡队列
// groupNumber 分组数量
func New(groupNumber int) *BalanceQueue {
	ret := &BalanceQueue{
		groups: make([]*group, groupNumber),
		tables: make([]*groupArray, 10), // 本身会自动扩容，初始值不是很重要
		pool:   make(map[Element]*group),
	}

	for i := 0; i < len(ret.tables); i++ {
		ret.tables[i] = &groupArray{}
	}
	// 初始化平衡数组，所有平衡队列容量为0
	for i := 0; i < len(ret.groups); i++ {
		ret.groups[i] = &group{queuePos: i}
		ret.tables[0].queue = append(ret.tables[0].queue, ret.groups[i])
	}
	return ret
}

func (q *BalanceQueue) String() string {
	buf := strings.Builder{}
	buf.WriteString("BalanceQueue:\n")
	buf.WriteString(fmt.Sprintf("分组数量: %v\n", len(q.groups)))
	for k, v := range q.tables {
		buf.WriteString(fmt.Sprintf("元素数量%v: 组数量%v ==>", k, len(v.queue)))
		for _, vv := range v.queue {
			buf.WriteString(fmt.Sprintf("%v ", len(vv.Array)))
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

func (q *BalanceQueue) Update() {
	if q.index == len(q.groups) {
		q.index = 0
	}
	for _, v := range q.groups[q.index].Array {
		v.BalanceQueueHandler()
	}
	q.index++
}

func (q *BalanceQueue) Push(e Element) {
	if e == nil {
		return
	}

	if _, ok := q.pool[e]; ok {
		return
	}

	for k, v := range q.tables {
		size := len(v.queue)
		if size == 0 {
			continue
		}

		arr := v.queue[size-1]
		if k+1 >= len(q.tables) {
			q.tables = append(q.tables, &groupArray{})
		}
		q.tables[k+1].queue = append(q.tables[k+1].queue, arr)
		q.tables[k].queue = v.queue[:size-1]
		arr.queuePos = len(q.tables[k+1].queue) - 1
		arr.Array = append(arr.Array, e)
		q.pool[e] = arr
		return
	}
	return
}

func (q *BalanceQueue) Pop(e Element) {
	group, ok := q.pool[e]
	if !ok {
		return
	}
	delete(q.pool, e)
	count := len(group.Array)
	for i := 0; i < count; i++ {
		if group.Array[i] == e {
			group.Array[i] = group.Array[count-1]
			group.Array = group.Array[:count-1]
			bqPos := group.queuePos
			queCount := len(q.tables[count].queue)
			q.tables[count].queue[bqPos] = q.tables[count].queue[queCount-1]
			q.tables[count].queue[bqPos].queuePos = bqPos
			q.tables[count].queue = q.tables[count].queue[:queCount-1]
			q.tables[count-1].queue = append(q.tables[count-1].queue, group)
			group.queuePos = len(q.tables[count-1].queue) - 1
			return
		}
	}
}
