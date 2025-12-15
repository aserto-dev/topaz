package ds

import (
	"net/http"

	cerr "github.com/aserto-dev/errors"

	"google.golang.org/grpc/codes"
)

//nolint:lll // single line readability more important.
var (
	ErrObjectNotFound                    = cerr.NewAsertoError("E20031", codes.NotFound, http.StatusNotFound, "object not found")
	ErrRelationNotFound                  = cerr.NewAsertoError("E20032", codes.NotFound, http.StatusNotFound, "relation not found")
	ErrPermissionNotFound                = cerr.NewAsertoError("E20033", codes.NotFound, http.StatusNotFound, "permission not found")
	ErrInvalidRequest                    = cerr.NewAsertoError("E20040", codes.InvalidArgument, http.StatusBadRequest, "invalid request instance")
	ErrInvalidArgumentObject             = cerr.NewAsertoError("E20041", codes.InvalidArgument, http.StatusBadRequest, "object invalid argument")
	ErrInvalidArgumentRelation           = cerr.NewAsertoError("E20042", codes.InvalidArgument, http.StatusBadRequest, "relation invalid argument")
	ErrInvalidArgumentObjectIdentifier   = cerr.NewAsertoError("E20043", codes.InvalidArgument, http.StatusBadRequest, "object identifier invalid argument")
	ErrInvalidArgumentRelationIdentifier = cerr.NewAsertoError("E20044", codes.InvalidArgument, http.StatusBadRequest, "relation identifier invalid argument")
	ErrInvalidArgumentObjectTypeSelector = cerr.NewAsertoError("E20045", codes.InvalidArgument, http.StatusBadRequest, "object type selector invalid argument")
	ErrNoCompleteObjectIdentifier        = cerr.NewAsertoError("E20050", codes.FailedPrecondition, http.StatusPreconditionFailed, "relation identifier no complete object identifier")
	ErrGraphDirectionality               = cerr.NewAsertoError("E20051", codes.InvalidArgument, http.StatusPreconditionFailed, "unable to determine graph directionality")
)
