package resetter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// fakeResettable is a Resetter that counts its resets and can fail on demand.
type fakeResettable struct {
	name  string
	err   error
	calls int
}

func (f *fakeResettable) Name() string {
	return f.name
}

func (f *fakeResettable) Reset() error {
	f.calls++
	return f.err
}

// newTestRPC builds the rpc service of a plugin whose registry was populated
// through the Collects callback, the same way endure populates it.
func newTestRPC(t *testing.T, resettables ...*fakeResettable) *rpc {
	t.Helper()

	p := &Plugin{}
	require.NoError(t, p.Init())

	collects := p.Collects()
	require.Len(t, collects, 1)

	for _, r := range resettables {
		collects[0].Callback(r)
	}

	svc, ok := p.RPC().(*rpc)
	require.True(t, ok)

	return svc
}

func TestRPCList(t *testing.T) {
	t.Run("empty registry", func(t *testing.T) {
		var list []string
		require.NoError(t, newTestRPC(t).List(true, &list))
		// net/rpc encodes nil and an empty slice differently, callers get a list
		require.NotNil(t, list)
		require.Empty(t, list)
	})

	t.Run("populated registry", func(t *testing.T) {
		svc := newTestRPC(t, &fakeResettable{name: "first"}, &fakeResettable{name: "second"})

		var list []string
		require.NoError(t, svc.List(true, &list))
		require.ElementsMatch(t, []string{"first", "second"}, list)
	})
}

func TestRPCResetKnownPlugin(t *testing.T) {
	target := &fakeResettable{name: "first"}
	bystander := &fakeResettable{name: "second"}
	svc := newTestRPC(t, target, bystander)

	var done bool
	require.NoError(t, svc.Reset("first", &done))
	require.True(t, done)
	require.Equal(t, 1, target.calls)
	require.Zero(t, bystander.calls)
}

func TestRPCResetUnknownPlugin(t *testing.T) {
	registered := &fakeResettable{name: "first"}
	svc := newTestRPC(t, registered)

	done := true
	err := svc.Reset("missing", &done)
	require.ErrorIs(t, err, errNoSuchPlugin)
	require.ErrorContains(t, err, "missing")
	require.False(t, done)
	require.Zero(t, registered.calls)
}

func TestRPCResetPluginError(t *testing.T) {
	resetErr := errors.New("pool allocation failed")
	failing := &fakeResettable{name: "first", err: resetErr}
	svc := newTestRPC(t, failing)

	done := true
	err := svc.Reset("first", &done)
	// the plugin adds no context, the resettable's error is passed through
	require.ErrorIs(t, err, resetErr)
	require.False(t, done)
	require.Equal(t, 1, failing.calls)
}
