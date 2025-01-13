package scaleout

import (
	"context"

	"github.com/goodplayer/scaleout/container"
)

type Configure struct {
	container.NopLifecycle

	ctx context.Context

	PrepareFn func(ctx context.Context) (context.Context, error)
	OnStartFn func(ctx context.Context) error
	OnStopFn  func(ctx context.Context) error
}

func (c *Configure) Prepare() error {
	if c.ctx == nil {
		c.ctx = context.Background()
	}
	newCtx, err := c.PrepareFn(c.ctx)
	if err != nil {
		return err
	}
	c.ctx = newCtx
	return nil
}

func (c *Configure) AfterStart() error {
	err := c.OnStartFn(c.ctx)
	if err != nil {
		return err
	}
	return nil
}

func (c *Configure) BeforeStop() error {
	err := c.OnStopFn(c.ctx)
	if err != nil {
		return err
	}
	return nil
}
