package resetter

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	resetterV1 "github.com/roadrunner-server/api-go/v6/resetter/v1"
)

var errNoSuchPlugin = errors.New("no such plugin")

type rpc struct {
	srv *Plugin
}

func (r *rpc) ListPlugins(_ *resetterV1.ListPluginsRequest, out *resetterV1.PluginsList) error {
	out.Plugins = slices.Collect(maps.Keys(r.srv.registry))
	return nil
}

func (r *rpc) Reset(in *resetterV1.ResetRequest, out *resetterV1.Response) error {
	name := in.GetPlugin()
	svc, ok := r.srv.registry[name]
	if !ok {
		return fmt.Errorf("%w: %s", errNoSuchPlugin, name)
	}
	if err := svc.Reset(); err != nil {
		return err
	}
	out.Ok = true
	return nil
}
