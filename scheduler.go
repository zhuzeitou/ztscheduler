package ztscheduler

import (
	"container/list"
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type Scheduler struct {
	async     bool
	ctx       context.Context
	cancel    context.CancelFunc
	runCtx    context.Context
	runCancel context.CancelFunc
	taskQueue list.List
	curId     int64
	mutex     sync.Mutex
	runMutex  sync.Mutex
}

type ztask struct {
	id       int64
	next     time.Time
	interval time.Duration // interval>0 means task is recurring
	fn       func()
}

func NewScheduler(async bool) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{async: async, ctx: ctx, cancel: cancel}
}

func (s *Scheduler) addTaskToQueue(task *ztask) {
	elem := s.taskQueue.Front()
__loop:
	if elem == nil {
		s.taskQueue.PushBack(task)
		if s.runCancel != nil {
			s.runCancel()
		}
		return
	}
	if task.next.Before(elem.Value.(*ztask).next) {
		s.taskQueue.InsertBefore(task, elem)
		if s.runCancel != nil {
			s.runCancel()
		}
		return
	}
	elem = elem.Next()
	goto __loop
}

func (s *Scheduler) reschedule(task *ztask) {
	if task.interval > 0 {
		task.next = task.next.Add(task.interval)
		s.addTaskToQueue(task)
	}
}

func (s *Scheduler) runTask(task *ztask) {
	if !s.async {
		s.runMutex.Lock()
		defer s.runMutex.Unlock()
	}
	task.fn()
}

func (s *Scheduler) check() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	now := time.Now()
__loop:
	if elem := s.taskQueue.Front(); elem == nil || now.Before(elem.Value.(*ztask).next) {
		return
	} else {
		task := s.taskQueue.Remove(elem).(*ztask)
		s.reschedule(task)
		go s.runTask(task)
		goto __loop
	}
}

func (s *Scheduler) run() {
__loop:
	s.mutex.Lock()
	if elem := s.taskQueue.Front(); elem == nil {
		s.runCtx, s.runCancel = context.WithCancel(s.ctx)
	} else {
		s.runCtx, s.runCancel = context.WithDeadline(s.ctx, elem.Value.(*ztask).next)
	}
	s.mutex.Unlock()
	select {
	case <-s.ctx.Done():
		return
	case <-s.runCtx.Done():
		s.check()
	}
	goto __loop
}

func (s *Scheduler) AddTask(delay time.Duration, fn func()) int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	next := time.Now().Add(delay)
	id := atomic.AddInt64(&s.curId, 1)
	task := &ztask{id: id, next: next, fn: fn}
	s.addTaskToQueue(task)
	return id
}

func (s *Scheduler) AddRecurringTask(delay time.Duration, interval time.Duration, fn func()) int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	next := time.Now().Add(delay)
	id := atomic.AddInt64(&s.curId, 1)
	task := &ztask{id: id, next: next, interval: interval, fn: fn}
	s.addTaskToQueue(task)
	return id
}

func (s *Scheduler) RemoveTask(id int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	elem := s.taskQueue.Front()
__loop:
	if elem == nil {
		return
	}
	if elem.Value.(*ztask).id == id {
		s.taskQueue.Remove(elem)
		return
	}
	elem = elem.Next()
	goto __loop
}

func (s *Scheduler) Start() {
	go s.run()
}

func (s *Scheduler) Stop() {
	s.cancel()
}

func (s *Scheduler) Done() <-chan struct{} {
	return s.ctx.Done()
}
