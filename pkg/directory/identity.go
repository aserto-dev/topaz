package directory

import (
	"context"
	"strings"

	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-directory/pkg/derr"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	User          string = "user"
	DummyUser     string = "topaz-user"
	Identifier    string = "identifier"
	Identity      string = "identity"
	DummyIdentity string = "topaz-identity"
	Reason        string = "reason"
)

type identityResolutionDirection int

const (
	// object_type:identity->subject_type:user (legacy).
	identityToUser identityResolutionDirection = iota + 1
	// object_type:user->subject_type:identity.
	userToIdentity
)

func resolutionDirection(ctx context.Context, client dsr3.ReaderClient) identityResolutionDirection {
	resp, _ := client.Check(ctx, &dsr3.CheckRequest{
		ObjectType:  User,
		ObjectId:    DummyUser,
		Relation:    Identifier,
		SubjectType: Identity,
		SubjectId:   DummyIdentity,
	})

	reason, ok := resp.GetContext().GetFields()[Reason]
	if ok && strings.HasPrefix(reason.GetStringValue(), derr.ErrRelationTypeNotFound.Code+" "+derr.ErrRelationTypeNotFound.Message) {
		return identityToUser
	}

	return userToIdentity
}

func ResolveIdentity(ctx context.Context, client dsr3.ReaderClient, identity string) (*dsc3.Object, error) {
	switch resolutionDirection(ctx, client) {
	// object_type:identity->subject_type:user (legacy).
	case identityToUser:
		return resolveIdentityLegacy(ctx, client, identity)

	// object_type:user->subject_type:identity.
	case userToIdentity:
		return resolveIdentity(ctx, client, identity)

	// fallback to validate both paths.
	default:
		if obj, err := resolveIdentity(ctx, client, identity); err == nil {
			return obj, nil
		}

		return resolveIdentityLegacy(ctx, client, identity)
	}
}

// resolveIdentity, resolves object_type:user->subject_type:identity (inverted identity).
func resolveIdentity(ctx context.Context, client dsr3.ReaderClient, identity string) (*dsc3.Object, error) {
	relReq := &dsr3.GetRelationRequest{
		ObjectType:  User,
		Relation:    Identifier,
		SubjectType: Identity,
		SubjectId:   identity,
		WithObjects: true,
	}

	return resolveIdentityToUser(ctx, client, relReq)
}

// resolveIdentityLegacy, resolves object_type:identity->subject_type:user (legacy).
func resolveIdentityLegacy(ctx context.Context, client dsr3.ReaderClient, identity string) (*dsc3.Object, error) {
	relReq := &dsr3.GetRelationRequest{
		ObjectType:  Identity,
		ObjectId:    identity,
		Relation:    Identifier,
		SubjectType: User,
		WithObjects: true,
	}

	return resolveIdentityToUser(ctx, client, relReq)
}

func resolveIdentityToUser(ctx context.Context, client dsr3.ReaderClient, relReq *dsr3.GetRelationRequest) (*dsc3.Object, error) {
	relResp, err := client.GetRelation(ctx, relReq)

	switch {
	case err != nil && status.Code(err) == codes.NotFound:
		return nil, aerr.ErrDirectoryObjectNotFound
	case err != nil:
		return nil, err
	case relResp.GetResult() == nil:
		return nil, aerr.ErrDirectoryObjectNotFound
	case len(relResp.GetObjects()) == 0:
		return &dsc3.Object{
			Type: User,
			Id: lo.Ternary(
				relResp.GetResult().GetObjectType() == User,
				relResp.GetResult().GetObjectId(),
				relResp.GetResult().GetSubjectId(),
			),
		}, nil
	}

	for k, v := range relResp.GetObjects() {
		if strings.HasPrefix(k, User+":") {
			return v, nil
		}
	}

	return nil, aerr.ErrDirectoryObjectNotFound
}
