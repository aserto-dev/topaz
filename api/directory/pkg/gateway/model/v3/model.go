package model

// import (
// 	"context"
// 	"io"
// 	"net/http"

// 	dms3 "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
// 	"github.com/aserto-dev/go-directory/pkg/manifest"
// 	"github.com/aserto-dev/go-directory/pkg/pb"
// 	"github.com/go-http-utils/headers"
// 	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
// 	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
// 	"github.com/pkg/errors"
// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/grpclog"
// )

// const MaxChunkSizeBytes int = 64 * 1024

// type metadataOption bool

// const (
// 	WithBody     metadataOption = false
// 	MetadataOnly metadataOption = true
// )

// func RegisterModelStreamHandlersFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
// 	conn, err := grpc.NewClient(endpoint, opts...)
// 	if err != nil {
// 		return err
// 	}

// 	defer func() {
// 		if err != nil {
// 			if cerr := conn.Close(); cerr != nil {
// 				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
// 			}

// 			return
// 		}

// 		go func() {
// 			<-ctx.Done()

// 			if cerr := conn.Close(); cerr != nil {
// 				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
// 			}
// 		}()
// 	}()

// 	return RegisterModelStreamHandlerClient(ctx, mux, dms3.NewModelClient(conn))
// }

// func RegisterModelStreamHandlerClient(ctx context.Context, mux *runtime.ServeMux, client dms3.ModelClient) error {
// 	if err := mux.HandlePath(
// 		"HEAD",
// 		"/api/v3/directory/manifest",
// 		getManifestHandler(mux, client, MetadataOnly),
// 	); err != nil {
// 		return errors.Wrap(err, "failed to register GetManifest handler")
// 	}

// 	if err := mux.HandlePath(
// 		"GET",
// 		"/api/v3/directory/manifest",
// 		getManifestHandler(mux, client, WithBody),
// 	); err != nil {
// 		return errors.Wrap(err, "failed to register GetManifest handler")
// 	}

// 	if err := mux.HandlePath(
// 		"POST",
// 		"/api/v3/directory/manifest",
// 		setManifestHandler(mux, client),
// 	); err != nil {
// 		return errors.Wrap(err, "failed to register SetManifest handler")
// 	}

// 	return nil
// }

// //nolint:gocognit,cyclop
// func getManifestHandler(mux *runtime.ServeMux, client dms3.ModelClient, mdOpt metadataOption) runtime.HandlerFunc {
// 	return func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
// 		_, outboundMarshaler := runtime.MarshalerForRequest(mux, req)

// 		ctx, err := runtime.AnnotateContext(
// 			req.Context(),
// 			mux,
// 			req,
// 			"/aserto.directory.model.v3.Model/GetManifest",
// 			runtime.WithHTTPPathPattern("/api/v3/directory/manifest"),
// 		)
// 		if err != nil {
// 			runtime.HTTPError(req.Context(), mux, outboundMarshaler, w, req, err)
// 			return
// 		}

// 		if mdOpt == MetadataOnly {
// 			md := metautils.ExtractOutgoing(ctx).Clone()
// 			md.Set(manifest.HeaderAsertoManifestRequest, "metadata-only")
// 			ctx = md.ToOutgoing(ctx)
// 		}

// 		stream, err := client.GetManifest(ctx, &dms3.GetManifestRequest{})
// 		if err != nil {
// 			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
// 			return
// 		}

// 		var (
// 			hasManifest bool
// 			hasBody     bool
// 			hasModel    bool
// 		)

// 		for {
// 			msg, err := stream.Recv()
// 			if errors.Is(err, io.EOF) {
// 				break
// 			} else if err != nil {
// 				runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
// 				return
// 			}

// 			if md := msg.GetMetadata(); md != nil && md.GetEtag() != "" {
// 				w.Header().Set(manifest.HeaderAsertoUpdatedAt, md.GetUpdatedAt().AsTime().Format(http.TimeFormat))
// 				w.Header().Set(headers.ETag, md.GetEtag())

// 				hasManifest = true
// 			}

// 			if body := msg.GetBody(); body != nil {
// 				hasBody = true

// 				w.Header().Set(headers.ContentType, "application/yaml")

// 				if _, err := w.Write(body.GetData()); err != nil { //nolint:gosec // data comes from store which passed parser validation.
// 					runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
// 					return
// 				}
// 			}

// 			// We send either the body or the model, never both.
// 			if model := msg.GetModel(); model != nil && !hasBody {
// 				hasModel = true

// 				w.Header().Set(headers.ContentType, "application/json")

// 				if err := pb.ProtoToBuf(w, model); err != nil {
// 					runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
// 					return
// 				}
// 			}
// 		}

// 		if hasManifest && !hasBody && !hasModel && req.Header.Get(headers.IfNoneMatch) != "" {
// 			w.WriteHeader(http.StatusNotModified)
// 		}
// 	}
// }

// func setManifestHandler(mux *runtime.ServeMux, client dms3.ModelClient) runtime.HandlerFunc {
// 	return func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
// 		_, outboundMarshaler := runtime.MarshalerForRequest(mux, req)

// 		ctx, err := runtime.AnnotateContext(
// 			req.Context(),
// 			mux,
// 			req,
// 			"/aserto.directory.model.v3.Model/SetManifest",
// 			runtime.WithHTTPPathPattern("/api/v3/directory/manifest"),
// 		)
// 		if err != nil {
// 			runtime.HTTPError(req.Context(), mux, outboundMarshaler, w, req, err)
// 			return
// 		}

// 		stream, err := client.SetManifest(ctx)
// 		if err != nil {
// 			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
// 			return
// 		}

// 		reader := req.Body

// 		defer func() { _ = reader.Close() }()

// 		buf := make([]byte, MaxChunkSizeBytes)

// 		for {
// 			n, err := reader.Read(buf)
// 			if err != nil && err != io.EOF {
// 				runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
// 				return
// 			}

// 			if n > 0 {
// 				if err := stream.Send(&dms3.SetManifestRequest{
// 					Msg: &dms3.SetManifestRequest_Body{Body: &dms3.Body{Data: buf[:n]}},
// 				}); err != nil {
// 					runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
// 					return
// 				}
// 			}

// 			if err == io.EOF {
// 				break
// 			}
// 		}

// 		if _, err := stream.CloseAndRecv(); err != nil {
// 			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
// 			return
// 		}
// 	}
// }
