package manifest

// import (
// 	"context"

// 	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
// 	"github.com/samber/lo"
// )

// const (
// 	HeaderAsertoManifestRequest = "Aserto-Manifest-Request"
// 	HeaderAsertoUpdatedAt       = "Aserto-Updated-At"
// )

// type ManifestRequest string

// const (
// 	// ManifestRequestDefault return the manifest metadata and body.
// 	ManifestRequestDefault ManifestRequest = ""
// 	// ManifestRequestMetadataOnly only return the manifest metadata.
// 	ManifestRequestMetadataOnly ManifestRequest = "metadata-only"
// 	// ManifestRequestModelOnly only return the manifest metadata and model.
// 	ManifestRequestModelOnly ManifestRequest = "model-only"
// 	// ManifestRequestWithModel return the manifest metadata, body, and model.
// 	ManifestRequestWithModel ManifestRequest = "with-model"
// )

// func IncomingManifestRequest(ctx context.Context) ManifestRequest {
// 	md := metautils.ExtractIncoming(ctx)
// 	amr := ManifestRequest(md.Get(HeaderAsertoManifestRequest))

// 	if !lo.Contains([]ManifestRequest{ManifestRequestMetadataOnly, ManifestRequestModelOnly, ManifestRequestWithModel}, amr) {
// 		amr = ManifestRequestDefault
// 	}

// 	return amr
// }

// func (mr ManifestRequest) WithBody() bool {
// 	return lo.Contains([]ManifestRequest{ManifestRequestDefault, ManifestRequestWithModel}, mr)
// }

// func (mr ManifestRequest) WithModel() bool {
// 	return lo.Contains([]ManifestRequest{ManifestRequestModelOnly, ManifestRequestWithModel}, mr)
// }
