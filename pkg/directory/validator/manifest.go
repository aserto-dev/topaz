package validator

import (
	"errors"

	dsm "github.com/aserto-dev/topaz/api/directory/v3/model"
	"github.com/aserto-dev/topaz/pkg/directory/gateway/model/v3"
)

func GetManifestRequest(msg *dsm.GetManifestRequest) error {
	return nil
}

func SetManifestRequest(msg *dsm.SetManifestRequest) error {
	return nil
}

func DeleteManifestRequest(msg *dsm.DeleteManifestRequest) error {
	return nil
}

func Metadata(msg *dsm.Metadata) error {
	return nil
}

var ErrBodyDataSize = errors.New("data size exceeds max chunk size of 65536 bytes")

func Body(msg *dsm.Body) error {
	if msg == nil {
		return nil
	}

	if len(msg.GetData()) > model.MaxChunkSizeBytes {
		return ErrBodyDataSize
	}

	return nil
}
