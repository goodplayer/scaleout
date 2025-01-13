package container

import "sort"

type Lifecycle interface {
	Order() int
	Prepare() error
	BeforeStart() error
	AfterStart() error
	BeforeStop() error
	AfterStop() error
}

type NopLifecycle struct {
}

func (n *NopLifecycle) Order() int {
	return 0
}

func (n *NopLifecycle) Prepare() error {
	return nil
}

func (n *NopLifecycle) BeforeStart() error {
	return nil
}

func (n *NopLifecycle) AfterStart() error {
	return nil
}

func (n *NopLifecycle) BeforeStop() error {
	return nil
}

func (n *NopLifecycle) AfterStop() error {
	return nil
}

type Container struct {
	lifecycles []Lifecycle
	startedIdx int
}

func NewContainer() *Container {
	return &Container{
		startedIdx: -1,
	}
}

func (c *Container) AddLifecycle(l Lifecycle) {
	c.lifecycles = append(c.lifecycles, l)
	sort.Slice(c.lifecycles, func(i, j int) bool {
		return c.lifecycles[i].Order() >= c.lifecycles[j].Order()
	})
}

func (c *Container) Startup() error {
	for c.startedIdx+1 < len(c.lifecycles) {
		if err := c.lifecycles[c.startedIdx+1].Prepare(); err != nil {
			return err
		}
		c.startedIdx++
	}
	for _, l := range c.lifecycles {
		if err := l.BeforeStart(); err != nil {
			return err
		}
	}
	for _, l := range c.lifecycles {
		if err := l.AfterStart(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Container) Shutdown() error {
	if c.startedIdx < 0 {
		return nil
	}
	for i := 0; i <= c.startedIdx; i++ {
		if err := c.lifecycles[i].BeforeStop(); err != nil {
			return err
		}
	}
	for i := 0; i < c.startedIdx; i++ {
		if err := c.lifecycles[i].AfterStop(); err != nil {
			return err
		}
	}
	return nil
}
