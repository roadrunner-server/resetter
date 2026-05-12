package resetter

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	resetterV1 "github.com/roadrunner-server/api-go/v6/resetter/v1"
	"github.com/roadrunner-server/api-go/v6/resetter/v1/resetterV1connect"
	"github.com/roadrunner-server/config/v6"
	"github.com/roadrunner-server/endure/v2"
	"github.com/roadrunner-server/logger/v6"
	"github.com/roadrunner-server/resetter/v6"
	rpcPlugin "github.com/roadrunner-server/rpc/v6"
	"github.com/roadrunner-server/server/v6"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const resetterAPIAddr = "127.0.0.1:6001"

// startResetterAPIContainer brings up rpc + server + logger + resetter +
// Plugin1 so the wire-surface tests below have a real resettable plugin
// registered. Plugin1.Name() returns "resetter.plugin1" — that's the only
// name visible via ListPlugins / Reset in this container.
func startResetterAPIContainer(t *testing.T) func() {
	t.Helper()

	cont := endure.New(slog.LevelError)
	cfg := &config.Plugin{
		Version: "2024.2.0",
		Path:    ".rr-resetter-api.yaml",
	}

	require.NoError(t, cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		&resetter.Plugin{},
		&Plugin1{},
	))
	require.NoError(t, cont.Init())

	ch, err := cont.Serve()
	require.NoError(t, err)

	wg := &sync.WaitGroup{}
	stop := make(chan struct{})
	wg.Go(func() {
		select {
		case e := <-ch:
			t.Errorf("container reported error: %v", e.Error)
		case <-stop:
		}
	})

	time.Sleep(500 * time.Millisecond)

	return func() {
		close(stop)
		require.NoError(t, cont.Stop())
		wg.Wait()
	}
}

func TestResetterConnectAPI(t *testing.T) {
	stop := startResetterAPIContainer(t)
	defer stop()

	httpc := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				return (&net.Dialer{Timeout: 30 * time.Second}).DialContext(ctx, network, addr)
			},
		},
	}
	client := resetterV1connect.NewResetterServiceClient(httpc, "http://"+resetterAPIAddr)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	listResp, err := client.ListPlugins(ctx, connect.NewRequest(&resetterV1.ListPluginsRequest{}))
	require.NoError(t, err)
	require.Contains(t, listResp.Msg.GetPlugins(), "resetter.plugin1")

	resetResp, err := client.Reset(ctx, connect.NewRequest(&resetterV1.ResetRequest{Plugin: "resetter.plugin1"}))
	require.NoError(t, err)
	require.True(t, resetResp.Msg.GetOk())

	// negative path: unknown plugin name must surface as CodeNotFound
	// (not the default CodeInternal) so clients can distinguish bad input
	// from real server faults.
	_, err = client.Reset(ctx, connect.NewRequest(&resetterV1.ResetRequest{Plugin: "does-not-exist"}))
	require.Error(t, err)
	require.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}

// TestResetterHTTPApi exercises both RPCs through plain HTTP/1.1 with a
// protojson body — the shape any non-Connect HTTP client uses against this
// handler.
func TestResetterHTTPApi(t *testing.T) {
	stop := startResetterAPIContainer(t)
	defer stop()

	httpc := &http.Client{Timeout: 30 * time.Second}
	ctx := t.Context()

	call := func(method string, in proto.Message, out proto.Message) {
		body, err := protojson.Marshal(in)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost,
			"http://"+resetterAPIAddr+"/resetter.v1.ResetterService/"+method, bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpc.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equalf(t, http.StatusOK, resp.StatusCode, "method=%s body=%s", method, respBody)
		require.NoError(t, protojson.Unmarshal(respBody, out))
	}

	var listResp resetterV1.PluginsList
	call("ListPlugins", &resetterV1.ListPluginsRequest{}, &listResp)
	require.Contains(t, listResp.GetPlugins(), "resetter.plugin1")

	var resetResp resetterV1.Response
	call("Reset", &resetterV1.ResetRequest{Plugin: "resetter.plugin1"}, &resetResp)
	require.True(t, resetResp.GetOk())
}

// TestResetterGRPCApi exercises both RPCs through a regular gRPC client. The
// same Connect handler serves gRPC framing off the same port.
func TestResetterGRPCApi(t *testing.T) {
	stop := startResetterAPIContainer(t)
	defer stop()

	conn, err := grpc.NewClient(resetterAPIAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	client := resetterV1.NewResetterServiceClient(conn)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	listResp, err := client.ListPlugins(ctx, &resetterV1.ListPluginsRequest{})
	require.NoError(t, err)
	require.Contains(t, listResp.GetPlugins(), "resetter.plugin1")

	resetResp, err := client.Reset(ctx, &resetterV1.ResetRequest{Plugin: "resetter.plugin1"})
	require.NoError(t, err)
	require.True(t, resetResp.GetOk())
}

// TestResetterHTTPGetIdempotency verifies that ListPlugins (marked
// NO_SIDE_EFFECTS) accepts HTTP GET, while Reset (mutating) returns 405.
func TestResetterHTTPGetIdempotency(t *testing.T) {
	stop := startResetterAPIContainer(t)
	defer stop()

	body, err := protojson.Marshal(&resetterV1.ResetRequest{Plugin: "probe"})
	require.NoError(t, err)

	q := url.Values{}
	q.Set("encoding", "json")
	q.Set("base64", "1")
	q.Set("message", base64.URLEncoding.EncodeToString(body))

	cases := []struct {
		method      string
		expectAllow bool
	}{
		{"ListPlugins", true},
		{"Reset", false},
	}

	httpc := &http.Client{Timeout: 30 * time.Second}
	for _, c := range cases {
		t.Run(c.method, func(t *testing.T) {
			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet,
				"http://"+resetterAPIAddr+"/resetter.v1.ResetterService/"+c.method+"?"+q.Encode(), nil)
			require.NoError(t, err)

			resp, err := httpc.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if c.expectAllow {
				require.NotEqualf(t, http.StatusMethodNotAllowed, resp.StatusCode,
					"%s via GET should be allowed; got 405\n%s", c.method, respBody)
				return
			}
			require.Equalf(t, http.StatusMethodNotAllowed, resp.StatusCode,
				"%s via GET should be rejected; got %s\n%s", c.method, resp.Status, respBody)
		})
	}
}
