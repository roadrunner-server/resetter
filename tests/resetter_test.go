package resetter

import (
	"log/slog"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	mocklogger "tests/mock"

	resetterV1 "github.com/roadrunner-server/api-go/v6/resetter/v1"
	"github.com/roadrunner-server/config/v6"
	"github.com/roadrunner-server/endure/v2"
	goridgeRpc "github.com/roadrunner-server/goridge/v4/pkg/rpc"
	"github.com/roadrunner-server/resetter/v6"
	rpcPlugin "github.com/roadrunner-server/rpc/v6"
	"github.com/roadrunner-server/server/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResetterInit(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2024.2.0",
		Path:    ".rr-resetter.yaml",
	}

	l, oLogger := mocklogger.SlogTestLogger(slog.LevelDebug)
	err := cont.RegisterAll(
		cfg,
		&server.Plugin{},
		l,
		&resetter.Plugin{},
		&rpcPlugin.Plugin{},
		&Plugin1{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	stopCh := make(chan struct{}, 1)

	wg := &sync.WaitGroup{}

	wg.Go(func() {
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	})

	time.Sleep(time.Second)

	t.Run("ResetterRpcTest", resetterRPCTest)
	stopCh <- struct{}{}
	wg.Wait()

	require.Equal(t, 1, oLogger.FilterMessageSnippet("plugin was started").Len())
}

func resetterRPCTest(t *testing.T) {
	conn, err := (&net.Dialer{}).DialContext(t.Context(), "tcp", "127.0.0.1:6001")
	require.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	defer func() { _ = client.Close() }()

	var resetResp resetterV1.Response
	err = client.Call("resetter.Reset", &resetterV1.ResetRequest{Plugin: "resetter.plugin1"}, &resetResp)
	assert.NoError(t, err)
	assert.True(t, resetResp.GetOk())

	// negative path: unknown plugin name must surface as an error over goridge net/rpc
	var missingResp resetterV1.Response
	err = client.Call("resetter.Reset", &resetterV1.ResetRequest{Plugin: "resetter.unknown"}, &missingResp)
	require.ErrorContains(t, err, "no such plugin")
	assert.False(t, missingResp.GetOk())

	var listResp resetterV1.PluginsList
	err = client.Call("resetter.ListPlugins", &resetterV1.ListPluginsRequest{}, &listResp)
	assert.NoError(t, err)
	require.Contains(t, listResp.GetPlugins(), "resetter.plugin1")
}
