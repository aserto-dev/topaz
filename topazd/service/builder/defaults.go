package builder

import (
	"net/http"

	"github.com/go-http-utils/headers"
)

var DefaultGatewayAllowedHeaders = []string{
	headers.Authorization,
	headers.ContentType,
	headers.IfMatch,
	headers.IfNoneMatch,
	"Depth",
}

var DefaultGatewayAllowedMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodHead,
	http.MethodDelete,
	http.MethodPut,
	http.MethodPatch,
	"PROPFIND",
	"MKCOL",
	"COPY",
	"MOVE",
}

var DefaultGatewayAllowedOrigins = []string{
	"http://localhost",
	"http://localhost:*",
	"https://localhost",
	"https://localhost:*",
	"http://127.0.0.1",
	"http://127.0.0.1:*",
	"https://127.0.0.1",
	"https://127.0.0.1:*",
}
