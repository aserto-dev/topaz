package aerr

import (
	"net/http"

	"github.com/aserto-dev/topaz/pkg/errors"
	"google.golang.org/grpc/codes"
)

var (
	// ErrRuntimeLoading the asked-for runtime is not yet available, but will likely be in the future.
	ErrRuntimeLoading = newErr("E30003", codes.Unavailable, http.StatusTooEarly, "runtime has not yet loaded")
	// ErrPolicyNotFound returned if a policy id is not found in the database.
	ErrPolicyNotFound = newErr("E10004", codes.NotFound, http.StatusNotFound, "policy not found")
	// ErrInvalidPolicyID returned if a policy id is invalid.
	ErrInvalidPolicyID = newErr("E30005", codes.InvalidArgument, http.StatusBadRequest, "invalid policy id")
	// ErrUserAlreadyExists returned if a user already exists.
	ErrUserAlreadyExists = newErr("E30006", codes.AlreadyExists, http.StatusConflict, "user already exists")
	// ErrAuthenticationFailed returned when authentication has failed or is not possible.
	ErrAuthenticationFailed = newErr("E30007", codes.Unauthenticated, http.StatusUnauthorized, "authentication failed")
	// ErrInvalidArgument returned when a given parameter is incorrect (wrong format, value or type).
	ErrInvalidArgument = newErr("E30008", codes.InvalidArgument, http.StatusBadRequest, "invalid argument")
	// ErrUserNotFound return if a user is not found.
	ErrUserNotFound = newErr("E30009", codes.NotFound, http.StatusNotFound, "user not found")
	// ErrBadQuery returned when a runtime query has an error.
	ErrBadQuery = newErr("E30010", codes.InvalidArgument, http.StatusBadRequest, "invalid query")
	// ErrInvalidDecision returned when a decision is invalid.
	ErrInvalidDecision = newErr("E30011", codes.InvalidArgument, http.StatusBadRequest, "invalid decision")
	// ErrBadRuntime returned when a runtime failed to load.
	ErrBadRuntime = newErr("E30012", codes.Unavailable, http.StatusServiceUnavailable, "runtime loading failed")
	// ErrDirectoryObjectNotFound returned if object id is not found in the directory.
	ErrDirectoryObjectNotFound = newErr("E30013", codes.NotFound, http.StatusNotFound, "directory object not found")
	// ErrInvalidPolicy returned if the loaded policy is invalid.
	ErrInvalidPolicy = newErr("E30014", codes.Internal, http.StatusInternalServerError, "invalid policy")
	// ErrAuthorizationFailed returned when authorization has failed or is not possible.
	ErrAuthorizationFailed = newErr("E30015", codes.PermissionDenied, http.StatusUnauthorized, "authorization failed")
	// ErrResolvingSubjectFromIdentityContext.
	ErrResolvingSubjectFromIdentityContext = newErr("E30016", codes.NotFound, http.StatusNotFound, "failed to resolve subject from identity context")
	// ErrResolvingUserFromSubject.
	ErrResolvingUserFromSubject = newErr("E30017", codes.NotFound, http.StatusNotFound, "failed to resolve user from subject")
)

func newErr(code string, statusCode codes.Code, httpCode int, msg string) *errors.AsertoError {
	return errors.NewAsertoError(code, statusCode, httpCode, msg)
}
