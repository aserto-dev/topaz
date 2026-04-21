package errors

import (
	"context"
	"net/http"
	"strconv"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

const (
	HTTPStatusErrorMetadata = "aserto-http-statuscode"
)

func CustomErrorHandler(
	ctx context.Context,
	gtw *runtime.ServeMux,
	runtimeMarshaler runtime.Marshaler,
	httpResponseWriter http.ResponseWriter,
	httpRequest *http.Request,
	err error,
) {
	if err == nil {
		runtime.DefaultHTTPErrorHandler(ctx, gtw, runtimeMarshaler, httpResponseWriter, httpRequest, err)
	}

	st := status.Convert(err)
	for _, detail := range st.Details() {
		errInfo, isErrInfo := detail.(*errdetails.ErrorInfo)
		if !isErrInfo {
			continue
		}

		value, hasErrorMetadata := errInfo.GetMetadata()[HTTPStatusErrorMetadata]
		if !hasErrorMetadata {
			continue
		}

		code, conversionErr := strconv.Atoi(value)
		if conversionErr != nil {
			logger := zerolog.Ctx(ctx)
			logger.Error().Err(conversionErr).Msg("Failed to detect http status code associated with this AsertoError")
		} else {
			var httpStatusError runtime.HTTPStatusError

			httpStatusError.Err = err
			httpStatusError.HTTPStatus = code
			runtime.DefaultHTTPErrorHandler(ctx, gtw, runtimeMarshaler, httpResponseWriter, httpRequest, &httpStatusError)

			return
		}
	}

	runtime.DefaultHTTPErrorHandler(ctx, gtw, runtimeMarshaler, httpResponseWriter, httpRequest, err)
}
