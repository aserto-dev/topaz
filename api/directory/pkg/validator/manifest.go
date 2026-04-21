package validator

// import (
// 	"errors"

// 	dsm3 "go-directory/aserto/directory/model/v3"

// 	"github.com/aserto-dev/go-directory/pkg/gateway/model/v3"
// )

// func GetManifestRequest(msg *dsm3.GetManifestRequest) error {
// 	return nil
// }

// func SetManifestRequest(msg *dsm3.SetManifestRequest) error {
// 	return nil
// }

// func DeleteManifestRequest(msg *dsm3.DeleteManifestRequest) error {
// 	return nil
// }

// func Metadata(msg *dsm3.Metadata) error {
// 	return nil
// }

// var ErrBodyDataSize = errors.New("data size exceeds max chunk size of 65536 bytes")

// func Body(msg *dsm3.Body) error {
// 	if msg == nil {
// 		return nil
// 	}

// 	if len(msg.GetData()) > model.MaxChunkSizeBytes {
// 		return ErrBodyDataSize
// 	}

// 	return nil
// }
