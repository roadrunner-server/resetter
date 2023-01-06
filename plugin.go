package resetter

import (
	"github.com/roadrunner-server/endure/v2/dep"
	"github.com/roadrunner-server/errors"
)

const PluginName = "resetter"

// Resetter interface
type Resetter interface {
	// Reset reload plugin
	Reset() error
	// Name of the plugin
	Name() string
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

// Collects declares services to be collected.
func (p *Plugin) Collects() []*dep.In {
	return []*dep.In{
		dep.Fits(func(pl any) {
			res := pl.(Resetter)
			p.registry[res.Name()] = res
		}, (*Resetter)(nil)),
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
