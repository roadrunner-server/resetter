package resetter

import (
	"errors"
	"fmt"
)

var errNoSuchPlugin = errors.New("no such plugin")

type rpc struct {
	srv *Plugin
}

// List all resettable plugins.
func (r *rpc) List(_ bool, list *[]string) error {
	*list = make([]string, 0, len(r.srv.registry))

	for name := range r.srv.registry {
		*list = append(*list, name)
	}

	return nil
}

// Reset named plugin.
func (r *rpc) Reset(service string, done *bool) error {
	svc, ok := r.srv.registry[service]
	if !ok {
		*done = false
		return fmt.Errorf("%w: %s", errNoSuchPlugin, service)
	}

	if err := svc.Reset(); err != nil {
		*done = false
		return err
	}

	*done = true

	return nil
}
