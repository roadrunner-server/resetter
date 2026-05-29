package resetter

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/roadrunner-server/api-go/v6/resetter/v1/resetterV1connect"
	"github.com/roadrunner-server/endure/v2/dep"
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

// Collects declare services to be collected.
func (p *Plugin) Collects() []*dep.In {
	return []*dep.In{
		dep.Fits(func(pl any) {
			res, ok := pl.(Resetter)
			if !ok {
				slog.Warn("plugin does not implement Resetter interface, skipping", slog.String("type", fmt.Sprintf("%T", pl)))
				return
			}
			p.registry[res.Name()] = res
		}, (*Resetter)(nil)),
	}
}

// Name of the service.
func (p *Plugin) Name() string {
	return PluginName
}

// RPC returns the Connect-RPC handler mount for the resetter service.
func (p *Plugin) RPC() (string, http.Handler) {
	return resetterV1connect.NewResetterServiceHandler(&rpc{srv: p})
}
