package resetter

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPluginNameAndInit(t *testing.T) {
	p := &Plugin{}
	// the rpc service is exposed under this name, "resetter.List" and "resetter.Reset"
	require.Equal(t, "resetter", p.Name())

	require.NoError(t, p.Init())
	require.NotNil(t, p.registry)
	require.Empty(t, p.registry)
}

func TestCollectsRegistersResettables(t *testing.T) {
	p := &Plugin{}
	require.NoError(t, p.Init())

	collects := p.Collects()
	require.Len(t, collects, 1)
	require.Equal(t, reflect.TypeFor[Resetter](), collects[0].Type)

	first := &fakeResettable{name: "first"}
	second := &fakeResettable{name: "second"}
	collects[0].Callback(first)
	collects[0].Callback(second)

	require.Equal(t, map[string]Resetter{"first": first, "second": second}, p.registry)
}

func TestRPCReturnsResetterService(t *testing.T) {
	p := &Plugin{}
	require.NoError(t, p.Init())

	svc, ok := p.RPC().(*rpc)
	require.True(t, ok)

	// the service reads the live registry, so plugins collected after RPC() was
	// called are still reachable
	p.Collects()[0].Callback(&fakeResettable{name: "late"})

	var list []string
	require.NoError(t, svc.List(true, &list))
	require.Equal(t, []string{"late"}, list)
}
