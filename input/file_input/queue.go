package file_input

import (
	"container/list"
	"sync"
)

type Queue struct {
	list list.List
	lock sync.Mutex
}

func (q *Queue) Add(v interface{}) {
	q.lock.Lock()
	q.list.PushBack(v)
	q.lock.Unlock()
}

func (q *Queue) Get() interface{} {
	q.lock.Lock()
	v := q.list.Front()

	if v != nil {
		q.list.Remove(v)
	}

	q.lock.Unlock()

	if v == nil {
		return nil
	}

	return v.Value
}
