package topaz

import (
	"net/http"

	ui "github.com/aserto-dev/go-aserto-one-ui"
	openapi "github.com/aserto-dev/openapi-grpc/publish/authorizer"
	"github.com/aserto-dev/topaz/pkg/app/server"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func HTTPRoutes(logger *zerolog.Logger, cfg *config.Config, directoryResolver resolvers.DirectoryResolver) (server.HTTPRouteRegistrations, error) {
	asertoUIHandler, err := ui.AsertoOneUIHandler(openapi.Static(), "/openapi.json")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to setup an aserto-one-ui http handler")
	}

	return func(mux *http.ServeMux) {
		mux.Handle("/", asertoUIHandler)
	}, nil
}
