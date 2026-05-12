package resetter

import (
	"context"
	stderr "errors"
	"fmt"

	"connectrpc.com/connect"
	resetterV1 "github.com/roadrunner-server/api-go/v6/resetter/v1"
)

var errNoSuchPlugin = stderr.New("no such plugin")

type rpc struct {
	srv *Plugin
}

func (r *rpc) ListPlugins(_ context.Context, _ *connect.Request[resetterV1.ListPluginsRequest]) (*connect.Response[resetterV1.PluginsList], error) {
	plugins := make([]string, 0, len(r.srv.registry))
	for name := range r.srv.registry {
		plugins = append(plugins, name)
	}
	return connect.NewResponse(&resetterV1.PluginsList{Plugins: plugins}), nil
}

func (r *rpc) Reset(_ context.Context, req *connect.Request[resetterV1.ResetRequest]) (*connect.Response[resetterV1.Response], error) {
	name := req.Msg.GetPlugin()
	svc, ok := r.srv.registry[name]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("%w: %s", errNoSuchPlugin, name))
	}
	if err := svc.Reset(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&resetterV1.Response{Ok: true}), nil
}
