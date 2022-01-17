package resetter

import (
	"github.com/roadrunner-server/api/v2/plugins/resetter"
	endure "github.com/roadrunner-server/endure/pkg/container"
	"github.com/roadrunner-server/errors"
)

const PluginName = "resetter"

type Plugin struct {
	registry map[string]resetter.Resetter
}

func (p *Plugin) Init() error {
	p.registry = make(map[string]resetter.Resetter)
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
func (p *Plugin) RegisterTarget(name endure.Named, r resetter.Resetter) error {
	p.registry[name.Name()] = r
	return nil
}

// Collects declares services to be collected.
func (p *Plugin) Collects() []interface{} {
	return []interface{}{
		p.RegisterTarget,
	}
}

// Name of the service.
func (p *Plugin) Name() string {
	return PluginName
}

// RPC returns associated rpc service.
func (p *Plugin) RPC() interface{} {
	return &rpc{srv: p}
}
