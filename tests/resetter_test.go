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

	"github.com/roadrunner-server/config/v4"
	"github.com/roadrunner-server/endure/v2"
	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
	"github.com/roadrunner-server/resetter/v4"
	rpcPlugin "github.com/roadrunner-server/rpc/v4"
	"github.com/roadrunner-server/server/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestResetterInit(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    ".rr-resetter.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
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
	wg.Add(1)

	go func() {
		defer wg.Done()
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
	}()

	time.Sleep(time.Second)

	t.Run("ResetterRpcTest", resetterRPCTest)
	stopCh <- struct{}{}
	wg.Wait()

	require.Equal(t, 1, oLogger.FilterMessageSnippet("plugin was started").Len())
}

func resetterRPCTest(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	// WorkerList contains a list of workers.

	var ret bool
	err = client.Call("resetter.Reset", "resetter.plugin1", &ret)
	assert.NoError(t, err)
	assert.True(t, ret)
	ret = false

	var services []string
	err = client.Call("resetter.List", nil, &services)
	assert.NotNil(t, services)
	assert.NoError(t, err)
	if services[0] != "resetter.plugin1" {
		t.Fatal("no enough services")
	}
}
