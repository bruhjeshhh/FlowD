package worker

import (
	"context"
	"sync"
	"sync/atomic"
)

type Controller struct {
	pauseMu     sync.RWMutex
	paused      atomic.Bool
	pauseChan   chan struct{}
	unpauseChan chan struct{}
}

func NewController() *Controller {
	return &Controller{
		pauseChan:   make(chan struct{}),
		unpauseChan: make(chan struct{}),
	}
}

func (c *Controller) IsPaused() bool {
	return c.paused.Load()
}

func (c *Controller) Pause() {
	c.pauseMu.Lock()
	defer c.pauseMu.Unlock()
	if !c.paused.Load() {
		c.paused.Store(true)
		close(c.pauseChan)
	}
}

func (c *Controller) Resume() {
	c.pauseMu.Lock()
	defer c.pauseMu.Unlock()
	if c.paused.Load() {
		c.paused.Store(false)
		c.pauseChan = make(chan struct{})
		c.unpauseChan = make(chan struct{})
	}
}

func (c *Controller) WaitWhilePaused(ctx context.Context) error {
	c.pauseMu.RLock()
	if !c.paused.Load() {
		c.pauseMu.RUnlock()
		return nil
	}
	pauseChan := c.pauseChan
	c.pauseMu.RUnlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-pauseChan:
		return nil
	}
}
