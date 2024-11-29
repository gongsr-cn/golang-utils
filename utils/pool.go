package utils

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

type Pool struct {
	Num    int
	task   chan func()
	count  *atomic.Int64
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewPool(num int) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pool{
		Num:    num,
		task:   make(chan func(), num),
		ctx:    ctx,
		cancel: cancel,
		count:  &atomic.Int64{},
	}
}

func (p *Pool) Start() {
	go p.run()
}

func (p *Pool) Close() {
	p.cancel()
	p.wg.Wait()
	close(p.task)
	for f := range p.task {
		p.wg.Add(1)
		go func(f func()) {
			defer p.wg.Done()
			f()
		}(f)
	}
	p.wg.Wait()
}

func (p *Pool) Task(f func()) {
	p.task <- f
}

func (p *Pool) run() {
	for {
		select {
		case f, ok := <-p.task:
			if !ok {
				return
			}
			p.wg.Add(1)
			go func() {
				defer p.wg.Done()
				f()
			}()
		case <-p.ctx.Done():
			fmt.Println("pool ctx done")
			return
		}
	}
}
