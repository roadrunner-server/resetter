package tests

import (
	"testing"

	"tests/helpers"

	"github.com/roadrunner-server/resetter/v6"
	rpcPlugin "github.com/roadrunner-server/rpc/v6"
	"github.com/roadrunner-server/server/v6"
	"github.com/stretchr/testify/require"
)

const rpcAddr = "127.0.0.1:6341"

// The resetter plugin collects every plugin implementing Resetter and exposes
// them over rpc, addressed by plugin name.
func TestResetterRPC(t *testing.T) {
	withPool := &Plugin1{}
	noPool := &Plugin2{}

	rr := helpers.Start(t, ".rr-resetter.yaml", []any{
		&server.Plugin{},
		&resetter.Plugin{},
		&rpcPlugin.Plugin{},
		withPool,
		noPool,
	}, helpers.WithTCPProbe(rpcAddr))

	client := helpers.RPC(t, rpcAddr)

	t.Run("List", func(t *testing.T) {
		var services []string
		require.NoError(t, client.Call("resetter.List", nil, &services))
		require.ElementsMatch(t, []string{"resetter.plugin1", "resetter.plugin2"}, services)
	})

	t.Run("Reset", func(t *testing.T) {
		var done bool
		require.NoError(t, client.Call("resetter.Reset", "resetter.plugin1", &done))
		require.True(t, done)
		// the call is routed by name, the other resettable is left alone
		require.Zero(t, noPool.Resets())

		require.NoError(t, client.Call("resetter.Reset", "resetter.plugin2", &done))
		require.True(t, done)
		require.Equal(t, int64(1), noPool.Resets())
	})

	t.Run("ResetUnknownPlugin", func(t *testing.T) {
		var done bool
		err := client.Call("resetter.Reset", "resetter.unknown", &done)
		require.ErrorContains(t, err, "no such plugin")
		require.False(t, done)
	})

	require.Equal(t, 1, rr.Logs.FilterMessageSnippet("plugin was started").Len())
}
