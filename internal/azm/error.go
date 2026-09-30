package azm

import (
	"net/http"

	cerr "github.com/aserto-dev/errors"

	"google.golang.org/grpc/codes"
)

var ErrInvalidSchemaVersion = cerr.NewAsertoError("E10000", codes.InvalidArgument, http.StatusBadRequest, "invalid or unsupported schema version")
