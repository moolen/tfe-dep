package orchestrator

import (
	"context"
	"sync"

	tfe "github.com/hashicorp/go-tfe"
)

type Orchestrator struct {
	client   *tfe.Client
	worker   int
	wg       *sync.WaitGroup
	stopCtx  context.Context
	shutdown context.CancelFunc
}

func New(client *tfe.Client, worker int) (*Orchestrator, error) {
	wg := &sync.WaitGroup{}
	stopCtx, shutdown := context.WithCancel(context.Background())
	o := &Orchestrator{
		client:   client,
		worker:   worker,
		wg:       wg,
		stopCtx:  stopCtx,
		shutdown: shutdown,
	}
	return o, nil
}

func (o *Orchestrator) Start() error {
	for i := 0; i < o.worker; i++ {
		o.wg.Add(1)
		go work(o.client, o.stopCtx, o.wg)
	}
	return nil
}

func (o *Orchestrator) Stop() error {
	o.shutdown()
	o.wg.Wait()
	return nil
}
