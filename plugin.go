package resetter

import (
	endure "github.com/roadrunner-server/endure/pkg/container"
	"github.com/roadrunner-server/errors"
)

const PluginName = "resetter"

// Resetter interface
type Resetter interface {
	// Reset reload plugin
	Reset() error
}

type Plugin struct {
	registry map[string]Resetter
}

func (p *Plugin) Init() error {
	p.registry = make(map[string]Resetter)
	return nil
}

// Reset named service.
func (p *Plugin) Reset(name string) error {
	const op = errors.Op("resetter_plugin_reset_by_name")
	svc, ok := p.registry[name]
	if !ok {
		return errors.E(op, errors.Errorf("no such plugin: %s", name))
	}

	return svc.Reset()
}

// RegisterTarget resettable service.
func (p *Plugin) RegisterTarget(name endure.Named, r Resetter) error {
	p.registry[name.Name()] = r
	return nil
}

// Collects declares services to be collected.
func (p *Plugin) Collects() []any {
	return []any{
		p.RegisterTarget,
	}
}

// Name of the service.
func (p *Plugin) Name() string {
	return PluginName
}

// RPC returns associated rpc service.
func (p *Plugin) RPC() any {
	return &rpc{srv: p}
}
