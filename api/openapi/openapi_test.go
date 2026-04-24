package openapi_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	openapi "github.com/aserto-dev/topaz/api/openapi"
	"github.com/samber/lo"
	req "github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestFilter(t *testing.T) {
	for _, services := range [][]string{
		{"reader"},
		{"writer"},
		{"model"},
		{"access"},
		{"reader", "writer"},
		{"model", "writer"},
		{"reader", "writer", "model", "assertion"},
	} {
		t.Run(strings.Join(services, ":"), func(tt *testing.T) {
			require := req.New(tt)

			b, err := openapi.Filter(openapi.Static(), services...)
			require.NoError(err)

			opIDs := lo.Map(gjson.GetBytes(b, "paths.@dig:operationId").Array(),
				func(v gjson.Result, _ int) string { return v.String() },
			)
			require.NotEmpty(opIDs)

			// Each remaining operation ID must match one of the expected services.
			for _, opID := range opIDs {
				require.True(openapi.MatchAny(opID, services...))
			}

			// Each of the expected services must match at least one operation ID.
			for _, svc := range services {
				require.True(hasServiceMatch(svc, opIDs...))
			}
		})
	}
}

func TestHandler(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(openapi.OpenAPIHandler("3030", "reader")))
	t.Cleanup(server.Close)

	require := req.New(t)

	resp, err := http.Get(server.URL) //nolint:noctx
	require.NoError(err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	body, err := io.ReadAll(resp.Body)
	require.NoError(err)

	t.Logf("Response: %s", body)

	opIDs := lo.Map(gjson.GetBytes(body, "paths.@dig:operationId").Array(),
		func(v gjson.Result, _ int) string { return v.String() },
	)
	require.NotEmpty(opIDs)

	for _, opID := range opIDs {
		require.True(openapi.MatchAny(opID, "reader"))
	}
}

func hasServiceMatch(svc string, vals ...string) bool {
	for _, val := range vals {
		if openapi.MatchAny(val, svc) {
			return true
		}
	}

	return false
}
