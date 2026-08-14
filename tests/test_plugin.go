package tests

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/roadrunner-server/pool/v2/payload"
	"github.com/roadrunner-server/pool/v2/pool"
	staticPool "github.com/roadrunner-server/pool/v2/pool/static_pool"
	"github.com/roadrunner-server/pool/v2/worker"
)

// testPoolConfig is the pool Plugin1 allocates. The resetter plugin behaves the
// same for any worker count, so the pool stays small: every worker is respawned
// on each of the ten iterations of Plugin1.Reset.
func testPoolConfig() *pool.Config {
	return &pool.Config{
		NumWorkers:      2,
		MaxJobs:         100,
		AllocateTimeout: time.Second * 10,
		DestroyTimeout:  time.Second * 10,
		Supervisor: &pool.SupervisorConfig{
			WatchTick:       60 * time.Second,
			TTL:             1000 * time.Second,
			IdleTTL:         10 * time.Second,
			ExecTTL:         10 * time.Second,
			MaxWorkerMemory: 1000,
		},
	}
}

type Configurer interface {
	// UnmarshalKey takes a single key and unmarshal it into a Struct.
	UnmarshalKey(name string, out any) error
	// Has checks if a config section exists.
	Has(name string) bool
}

// Server creates workers for the application.
type Server interface {
	NewPool(ctx context.Context, cfg *pool.Config, env map[string]string, _ *slog.Logger) (*staticPool.Pool, error)
	NewWorker(ctx context.Context, env map[string]string) (*worker.Process, error)
}

type Pool interface {
	// Workers return a worker list associated with the pool.
	Workers() (workers []*worker.Process)
	// Exec payload
	Exec(ctx context.Context, p *payload.Payload, stopCh chan struct{}) (chan *staticPool.PExec, error)
	// Reset kills all workers inside the watcher and replaces with new
	Reset(ctx context.Context) error
	// Destroy all underlying stacks (but let them complete the task).
	Destroy(ctx context.Context)
}

// Plugin1 is a resettable backed by a real worker pool, so Reset exercises the
// path a production plugin takes.
type Plugin1 struct {
	config Configurer
	server Server

	p Pool
}

func (p1 *Plugin1) Init(cfg Configurer, server Server) error {
	p1.config = cfg
	p1.server = server
	return nil
}

func (p1 *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	var err error
	p1.p, err = p1.server.NewPool(context.Background(), testPoolConfig(), nil, nil)
	if err != nil {
		errCh <- err
		return errCh
	}
	return errCh
}

func (p1 *Plugin1) Stop(context.Context) error {
	return nil
}

func (p1 *Plugin1) Name() string {
	return "resetter.plugin1"
}

func (p1 *Plugin1) Reset() error {
	for range 10 {
		err := p1.p.Reset(context.Background())
		if err != nil {
			return err
		}
	}

	return nil
}

// Plugin2 is a resettable without a pool. It makes the registry hold more than
// one entry, so List and the name-based dispatch have something to distinguish.
type Plugin2 struct {
	resets atomic.Int64
}

func (p2 *Plugin2) Init() error {
	return nil
}

func (p2 *Plugin2) Name() string {
	return "resetter.plugin2"
}

func (p2 *Plugin2) Reset() error {
	p2.resets.Add(1)
	return nil
}

// Resets returns how many times the plugin was reset.
func (p2 *Plugin2) Resets() int64 {
	return p2.resets.Load()
}
