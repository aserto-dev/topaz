package builder

import (
	"context"
	"net/http"
	"strconv"

	cerr "github.com/aserto-dev/errors"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/samber/lo"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func gatewayMux(allowedHeaders []string) *runtime.ServeMux {
	headerSet := lo.SliceToMap(allowedHeaders, func(header string) (string, struct{}) {
		return header, struct{}{}
	})

	return runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(func(header string) (string, bool) {
			if _, ok := headerSet[header]; ok {
				return header, true
			}

			return runtime.DefaultHeaderMatcher(header)
		}),
		runtime.WithMarshalerOption(
			runtime.MIMEWildcard,
			&runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					Indent:          "  ",
					AllowPartial:    true,
					UseProtoNames:   true,
					EmitUnpopulated: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					AllowPartial:   true,
					DiscardUnknown: true,
				},
			},
		),
		runtime.WithMarshalerOption(
			"application/json+masked",
			&runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					Indent:        "  ",
					AllowPartial:  true,
					UseProtoNames: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					AllowPartial:   true,
					DiscardUnknown: true,
				},
			},
		),
		runtime.WithUnescapingMode(runtime.UnescapingModeAllExceptSlash),
		runtime.WithForwardResponseOption(forwardXHTTPCode),
		runtime.WithErrorHandler(cerr.CustomErrorHandler),
	)
}

func forwardXHTTPCode(ctx context.Context, w http.ResponseWriter, _ proto.Message) error {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return nil
	}

	headers := metautils.NiceMD(md.HeaderMD)

	// set http status code
	if xcode := headers.Get("x-http-code"); xcode != "" {
		code, err := strconv.Atoi(xcode)
		if err != nil {
			return err
		}
		// delete the headers to not expose any grpc-metadata in http response
		headers.Del("x-http-code")
		delete(w.Header(), "Grpc-Metadata-X-Http-Code")
		w.WriteHeader(code)
	}

	return nil
}
