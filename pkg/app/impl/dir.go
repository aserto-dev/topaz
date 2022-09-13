package impl

import (
	"context"
	"encoding/base64"
	"io"
	"strings"

	"github.com/aserto-dev/aserto-grpc/grpcutil"
	xds "github.com/aserto-dev/go-eds"
	"github.com/aserto-dev/go-eds/pkg/kvs"
	api "github.com/aserto-dev/go-grpc/aserto/api/v1"
	dir "github.com/aserto-dev/go-grpc/aserto/authorizer/directory/v1"
	"github.com/aserto-dev/go-lib/ids"
	"github.com/aserto-dev/go-utils/cerr"
	"github.com/aserto-dev/go-utils/opts"
	"github.com/aserto-dev/topaz/directory"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
)

type DirectoryServer struct {
	logger            *zerolog.Logger
	directoryResolver resolvers.DirectoryResolver
}

func NewDirectoryServer(logger *zerolog.Logger, directoryResolver resolvers.DirectoryResolver) (*DirectoryServer, error) {
	logger.Info().Msg("new directory service handler")
	newLogger := logger.With().Str("component", "api.dir").Logger()

	return &DirectoryServer{
		logger:            &newLogger,
		directoryResolver: directoryResolver,
	}, nil
}

// Gets a directory based on the provided context.
// If tenantID is not an empty string, it will be used to get the directory.
// This is useful for getting the directory when a Tenant ID header is not available in the context.
func (s *DirectoryServer) eds(ctx context.Context, tenantID string) (directory.Directory, error) {
	// TODO: this shouldn't assume an eds singleton at the other end
	// we have to make upstream OPA allow us to configure the compiler of the plugin manager
	// so we can register builtins that are not global.

	if tenantID != "" {
		return s.directoryResolver.GetDirectory(ctx, tenantID)
	}

	ctxTenantID := grpcutil.ExtractTenantID(ctx)
	if ctxTenantID == "" {
		return nil, cerr.ErrInvalidTenantID
	}

	d, err := s.directoryResolver.DirectoryFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get resolve directory")
	}
	return d, nil
}

func (s *DirectoryServer) ListUsers(ctx context.Context, req *dir.ListUsersRequest) (*dir.ListUsersResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	s.logger.Trace().Str("tenantID", tenantID).Interface("req", req).Msg("ListUsers")

	// default pagination settings
	if req.Page == nil {
		req.Page = &api.PaginationRequest{
			Size:  kvs.SingleResultSet, // TODO should be kvs.ServerSetPageSize
			Token: "",
		}
	}

	// default fields settings
	if req.Fields == nil {
		req.Fields = &api.Fields{
			Mask: &field_mask.FieldMask{},
		}
	}

	pageSize := kvs.PageSize(req.Page.Size)
	paths := s.validateMask(req.Fields.Mask)

	resp := dir.ListUsersResponse{}

	pageToken, err := base64.URLEncoding.DecodeString(req.Page.Token)
	if err != nil {
		s.logger.Error().Err(err).Str("token", req.Page.Token).Msg("DecodeString")
		pageToken = []byte("")
	}

	var (
		users []*api.User
		next  string
		total int32
	)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.ListUsersResponse{}, errors.Wrap(err, "failed to get directory")
	}

	optParam := xds.WithDeleted(req.GetDeleted())

	if req.Base {
		users, next, total, err = eds.ListUsers(tenantID, string(pageToken), pageSize, paths, optParam)
	} else {
		users, next, total, err = eds.ListUsersExt(tenantID, string(pageToken), pageSize, paths, optParam)
	}
	if err != nil {
		return &resp, errors.Wrap(err, "failed to list users")
	}

	nextToken := base64.URLEncoding.EncodeToString([]byte(next))

	resp.Results = users
	resp.Page = &api.PaginationResponse{
		NextToken:  nextToken,
		ResultSize: int32(len(users)),
		TotalSize:  total,
	}

	return &resp, nil
}

func (s *DirectoryServer) GetUser(ctx context.Context, req *dir.GetUserRequest) (*dir.GetUserResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	s.logger.Trace().Str("tenantID", tenantID).Interface("req", req).Msg("GetUser")

	if req.Id == "" {
		return &dir.GetUserResponse{}, cerr.ErrInvalidArgument.Msg("id not set")
	}

	var (
		user *api.User
		err  error
	)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.GetUserResponse{}, errors.Wrap(err, "failed to get directory")
	}

	if req.Base {
		user, err = eds.GetUser(tenantID, req.Id)
	} else {
		user, err = eds.GetUserExt(tenantID, req.Id)
	}
	if err != nil {
		if errors.Is(err, kvs.ErrPathNotFound) || errors.Is(err, kvs.ErrKeyNotFound) {
			return &dir.GetUserResponse{}, cerr.ErrUserNotFound.Err(err).Str("userid", req.Id)
		}
		return &dir.GetUserResponse{}, err
	}

	return &dir.GetUserResponse{
		Result: user,
	}, nil
}

func (s *DirectoryServer) CreateUser(ctx context.Context, req *dir.CreateUserRequest) (*dir.CreateUserResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	s.logger.Trace().Str("tenantID", tenantID).Interface("req", req).Msg("CreateUser")

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.CreateUserResponse{}, errors.Wrap(err, "failed to get directory")
	}

	newUser, err := eds.CreateUser(tenantID, req.User)
	if err != nil {
		if errors.Is(err, kvs.ErrKeyExists) {
			return &dir.CreateUserResponse{}, cerr.ErrUserAlreadyExists.Err(err).Str("user-email", req.User.Email)
		}
		return &dir.CreateUserResponse{}, errors.Wrap(err, "failed to create user")
	}

	return &dir.CreateUserResponse{
		Result: newUser,
	}, nil
}

func (s *DirectoryServer) UpdateUser(ctx context.Context, req *dir.UpdateUserRequest) (*dir.UpdateUserResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	s.logger.Trace().Str("tenantID", tenantID).Interface("req", req).Msg("UpdateUser")

	if req.Id == "" {
		return &dir.UpdateUserResponse{}, cerr.ErrInvalidArgument.Msg("id not set")
	}

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.UpdateUserResponse{}, errors.Wrap(err, "failed to get directory")
	}

	newUser, err := eds.UpdateUser(tenantID, req.User)
	if err != nil {
		return &dir.UpdateUserResponse{}, errors.Wrap(err, "failed to update user")
	}

	return &dir.UpdateUserResponse{
		Result: newUser,
	}, nil
}

func (s *DirectoryServer) DeleteUser(ctx context.Context, req *dir.DeleteUserRequest) (*dir.DeleteUserResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	s.logger.Trace().Str("tenantID", tenantID).Interface("req", req).Msg("DeleteUser")

	if req.Id == "" {
		return &dir.DeleteUserResponse{}, cerr.ErrInvalidArgument.Msg("id not set")
	}

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.DeleteUserResponse{}, errors.Wrap(err, "failed to get directory")
	}

	err = eds.DeleteUser(tenantID, req.Id)

	return &dir.DeleteUserResponse{
		Result: &emptypb.Empty{},
	}, err
}

func (s *DirectoryServer) GetIdentity(ctx context.Context, req *dir.GetIdentityRequest) (*dir.GetIdentityResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	s.logger.Trace().Str("tenantID", tenantID).Interface("req", req).Msg("GetIdentity")

	if req.Identity == "" {
		return &dir.GetIdentityResponse{}, cerr.ErrInvalidArgument.Msg("identity not set")
	}

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.GetIdentityResponse{}, errors.Wrap(err, "failed to get directory")
	}

	uid, err := eds.GetIdentity(tenantID, req.Identity)
	if err != nil {
		if errors.Is(err, kvs.ErrPathNotFound) || errors.Is(err, kvs.ErrKeyNotFound) {
			return &dir.GetIdentityResponse{}, cerr.ErrUserNotFound.Err(err).Str("identity", req.Identity)
		}
		return &dir.GetIdentityResponse{}, err
	}

	return &dir.GetIdentityResponse{
		Id: uid,
	}, nil
}

// validateMask checks if provided mask is validate.
func (s *DirectoryServer) validateMask(mask *fieldmaskpb.FieldMask) []string {
	if len(mask.Paths) > 0 && mask.Paths[0] == "" {
		return []string{}
	}

	var u *api.User

	mask.Normalize()

	if !mask.IsValid(u) {
		s.logger.Error().Msgf("field mask invalid %q", mask.Paths)
		return []string{}
	}

	return mask.Paths
}

// LoadUsers load user stream into edge directory. (GRPC-only)
func (s *DirectoryServer) LoadUsers(stream dir.Directory_LoadUsersServer) error {
	tenantID := grpcutil.ExtractTenantID(stream.Context())

	res := &dir.LoadUsersResponse{
		Received: 0,
		Created:  0,
		Updated:  0,
		Deleted:  0,
		Errors:   0,
	}

	eds, err := s.eds(stream.Context(), "")
	if err != nil {
		return errors.Wrap(err, "failed to get directory")
	}

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			s.logger.Debug().Interface("res", res).Msg("load user response")
			return stream.SendAndClose(res)
		} else if err != nil {
			s.logger.Err(err).Msg("cannot receive req")
			return stream.SendAndClose(res)
		}

		if u := req.GetUser(); u != nil {
			s.userHandler(eds, tenantID, u, res)

		} else if x := req.GetUserExt(); x != nil {
			s.userExtHandler(eds, tenantID, x, res)

		} else if du := req.GetDeleteUser(); du != nil {
			s.deleteUserHandler(eds, tenantID, du, res)

		} else if dc := req.GetDeleteConnection(); dc != nil {
			s.deleteConnectionHandler(eds, tenantID, dc, res)

		} else {
			panic(errors.Errorf("unhandled loaduser message type [%s]",
				req.ProtoReflect().Descriptor().FullName()),
			)
		}
	}
}

func (s *DirectoryServer) userHandler(eds directory.Directory, tenantID string, u *api.User, res *dir.LoadUsersResponse) {
	s.logger.Debug().Msgf("get user %s", u.Id)

	res.Received++

	_, isNew, err := eds.UpsertUser(tenantID, u)
	if err != nil {
		s.logger.Error().Err(err).Msgf("upsert user id:[%s]", u.Id)
		res.Errors++
		return
	}

	if isNew {
		res.Created++
	} else {
		res.Updated++
	}
}

func (s *DirectoryServer) userExtHandler(eds directory.Directory, tenantID string, x *api.UserExt, res *dir.LoadUsersResponse) {
	s.logger.Debug().Msgf("get user extension id:[%s]", x.Id)

	res.Received++

	if err := eds.UpdateUserExt(tenantID, x); err != nil {
		s.logger.Error().Err(err).Msgf("update user ext id:[%s]", x.Id)
		res.Errors++
		return
	}

	res.Updated++
}

func (s *DirectoryServer) deleteUserHandler(eds directory.Directory, tenantID string, du *api.DeleteUser, res *dir.LoadUsersResponse) {
	s.logger.Debug().Msgf("delete user pid:[%s]", du.Id)

	uid, err := eds.GetIdentity(tenantID, du.Id)
	if err != nil {
		s.logger.Error().Err(err).Msgf("get identity pid:[%s]", du.Id)
		res.Errors++
		return
	}

	if err := eds.DeleteUser(tenantID, uid); err != nil {
		s.logger.Error().Err(err).Msgf("delete user id:[%s]", uid)
		res.Errors++
		return
	}

	res.Deleted++
}

func (s *DirectoryServer) deleteConnectionHandler(eds directory.Directory, tenantID string, dc *api.DeleteConnection, res *dir.LoadUsersResponse) {
	s.logger.Debug().Msgf("delete connection id:[%s]", dc.ConnectionId)

	pageToken := ""
	pageSize := int32(10)
	paths := []string{}
	options := []opts.Param{
		xds.WithDeleted(true),
	}

	for {
		users, nextToken, _, err := eds.ListUsers(tenantID, pageToken, pageSize, paths, options...)
		if err != nil {
			s.logger.Error().Err(err).Msgf("list users tenant:[%s]", tenantID)
			break
		}

		for _, u := range users {
			if u.Metadata == nil || u.Metadata.ConnectionId == nil {
				continue
			}

			if strings.EqualFold(*u.Metadata.ConnectionId, dc.ConnectionId) {
				if err := eds.DeleteUser(tenantID, u.Id); err != nil {
					s.logger.Error().Err(err).Msgf("delete user id:[%s]", u.Id)
					res.Errors++
					continue
				}
				res.Deleted++
			}
		}

		if nextToken == "" {
			break
		}
		pageToken = nextToken
	}
}

// ListTenants returns tenant id collection for tenants in edge directory instance. (GRPC-only)
func (s *DirectoryServer) ListTenants(ctx context.Context, req *dir.ListTenantsRequest) (*dir.ListTenantsResponse, error) {
	result := &dir.ListTenantsResponse{}

	dirs, err := s.directoryResolver.ListDirectories(ctx)
	if err != nil {
		return result, errors.Wrap(err, "failed to list user directories")
	}

	resp := dir.ListTenantsResponse{
		Results: dirs,
	}

	return &resp, nil
}

// CreateTenant, if tenant does not exist, creates a tenant namespace inside EDS and returns a tenant id.
// If no id is provided, a tenant id will be generated using the tenant id generator logic
func (s *DirectoryServer) CreateTenant(ctx context.Context, req *dir.CreateTenantRequest) (*dir.CreateTenantResponse, error) {
	s.logger.Trace().Str("id", req.Id).Msg("CreateTenant")

	resp := dir.CreateTenantResponse{}

	// check id
	if req.Id != "" {
		if err := ids.CheckTenantID(req.Id); err != nil {
			return &resp, errors.Wrapf(err, "tenant id invalid [%s]", req.Id)
		}
	} else {
		var err error
		req.Id, err = ids.NewTenantID()
		if err != nil {
			return &resp, errors.Wrapf(err, "generating tenant id")
		}
		resp.Id = req.Id
	}

	eds, err := s.eds(ctx, req.Id)
	if err != nil {
		return &resp, errors.Wrap(err, "failed to get directory")
	}

	// create tenant namespace if not existing
	if err := eds.CreateTenant(req.Id); err != nil {
		return &resp, errors.Wrapf(err, "create tenant")
	}

	return &dir.CreateTenantResponse{
		Id: req.Id,
	}, nil
}

// DeleteTenant, if tenant exists, remove the tenant namespace inside EDS
func (s *DirectoryServer) DeleteTenant(ctx context.Context, req *dir.DeleteTenantRequest) (*dir.DeleteTenantResponse, error) {
	resp := dir.DeleteTenantResponse{Result: &emptypb.Empty{}}

	err := s.directoryResolver.RemoveDirectory(ctx, req.Id)
	if err != nil {
		return &resp, errors.Wrapf(err, "failed to remove user directory")
	}

	// remove tenant namespace when existing
	return &resp, nil
}

func (s *DirectoryServer) ListResources(ctx context.Context, req *dir.ListResourcesRequest) (*dir.ListResourcesResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	// default pagination settings
	if req.Page == nil {
		req.Page = &api.PaginationRequest{
			Size:  kvs.ServerSetPageSize,
			Token: "",
		}
	}

	pageSize := kvs.PageSize(req.Page.Size)
	pageToken, err := base64.URLEncoding.DecodeString(req.Page.Token)
	if err != nil {
		s.logger.Error().Err(err).Str("token", req.Page.Token).Msg("DecodeString")
		pageToken = []byte("")
	}

	resp := dir.ListResourcesResponse{}

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &resp, errors.Wrap(err, "failed to get directory")
	}

	results, next, total, err := eds.ListResources(tenantID, string(pageToken), pageSize)
	if err != nil {
		return &resp, errors.Wrapf(err, "list resources for tenant [%s]", tenantID)
	}

	nextToken := base64.URLEncoding.EncodeToString([]byte(next))

	resp.Results = results
	resp.Page = &api.PaginationResponse{
		NextToken:  nextToken,
		ResultSize: int32(len(results)),
		TotalSize:  total,
	}

	return &resp, nil
}

func (s *DirectoryServer) GetResource(ctx context.Context, req *dir.GetResourceRequest) (*dir.GetResourceResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	resp := dir.GetResourceResponse{}

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &resp, errors.Wrap(err, "failed to get directory")
	}

	value, err := eds.GetResource(tenantID, req.Key)
	if err != nil {
		return &resp, errors.Wrapf(err, "get resource for key [%s]", req.Key)
	}

	resp.Value = value

	return &resp, nil
}

func (s *DirectoryServer) SetResource(ctx context.Context, req *dir.SetResourceRequest) (*dir.SetResourceResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	resp := dir.SetResourceResponse{}

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &resp, errors.Wrap(err, "failed to get directory")
	}

	if err := eds.SetResource(tenantID, req.Key, req.Value); err != nil {
		return &resp, errors.Wrapf(err, "set resource for key [%s]", req.Key)
	}

	return &resp, nil
}

func (s *DirectoryServer) DeleteResource(ctx context.Context, req *dir.DeleteResourceRequest) (*dir.DeleteResourceResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	resp := dir.DeleteResourceResponse{}

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &resp, errors.Wrap(err, "failed to get directory")
	}

	if err := eds.DeleteResource(tenantID, req.Key); err != nil {
		return &resp, errors.Wrapf(err, "delete resource for key [%s]", req.Key)
	}

	return &resp, nil
}

func (s *DirectoryServer) GetValue(ctx context.Context, req *dir.GetValueRequest) (*dir.GetValueResponse, error) {
	resp := dir.GetValueResponse{}

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &resp, errors.Wrap(err, "failed to get directory")
	}

	value, err := eds.GetValue(req.Path, req.Key)
	if err != nil {
		return &resp, errors.Wrapf(err, "get value")
	}

	resp.Value = value

	return &resp, nil
}

const NoAppl = ""

func (s *DirectoryServer) GetUserProperties(
	ctx context.Context,
	req *dir.GetUserPropertiesRequest) (*dir.GetUserPropertiesResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.GetUserPropertiesResponse{}, errors.Wrap(err, "failed to get directory")
	}

	properties, err := eds.GetUserProperties(tenantID, req.Id, NoAppl)
	if err != nil {
		return &dir.GetUserPropertiesResponse{}, errors.Wrapf(err, "get properties")
	}

	return &dir.GetUserPropertiesResponse{
		Results: properties,
	}, nil
}

func (s *DirectoryServer) SetUserProperties(
	ctx context.Context,
	req *dir.SetUserPropertiesRequest) (*dir.SetUserPropertiesResponse, error) {

	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.SetUserPropertiesResponse{}, errors.Wrap(err, "failed to get directory")
	}

	return &dir.SetUserPropertiesResponse{}, eds.SetUserProperties(tenantID, req.Id, NoAppl, req.Properties, false)
}

func (s *DirectoryServer) SetUserProperty(
	ctx context.Context,
	req *dir.SetUserPropertyRequest) (*dir.SetUserPropertyResponse, error) {

	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.SetUserPropertyResponse{}, errors.Wrap(err, "failed to get directory")
	}

	return &dir.SetUserPropertyResponse{}, eds.SetUserProperty(tenantID, req.Id, NoAppl, req.Key, req.Value, false)
}

func (s *DirectoryServer) DeleteUserProperty(
	ctx context.Context,
	req *dir.DeleteUserPropertyRequest) (*dir.DeleteUserPropertyResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.DeleteUserPropertyResponse{}, errors.Wrap(err, "failed to get directory")
	}

	return &dir.DeleteUserPropertyResponse{}, eds.SetUserProperty(tenantID, req.Id, NoAppl, req.Key, structpb.NewNullValue(), true)
}

func (s *DirectoryServer) GetUserRoles(ctx context.Context, req *dir.GetUserRolesRequest) (*dir.GetUserRolesResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.GetUserRolesResponse{}, errors.Wrap(err, "failed to get directory")
	}

	roles, err := eds.GetUserRoles(tenantID, req.Id, NoAppl)
	if err != nil {
		return &dir.GetUserRolesResponse{}, errors.Wrapf(err, "get roles")
	}

	return &dir.GetUserRolesResponse{
		Results: roles,
	}, nil
}

func (s *DirectoryServer) SetUserRoles(ctx context.Context, req *dir.SetUserRolesRequest) (*dir.SetUserRolesResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.SetUserRolesResponse{}, errors.Wrap(err, "failed to get directory")
	}

	return &dir.SetUserRolesResponse{}, eds.SetUserRoles(tenantID, req.Id, NoAppl, req.Roles, false)
}

func (s *DirectoryServer) SetUserRole(ctx context.Context, req *dir.SetUserRoleRequest) (*dir.SetUserRoleResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.SetUserRoleResponse{}, errors.Wrap(err, "failed to get directory")
	}

	return &dir.SetUserRoleResponse{}, eds.SetUserRole(tenantID, req.Id, NoAppl, req.Role, false)
}

func (s *DirectoryServer) DeleteUserRole(
	ctx context.Context,
	req *dir.DeleteUserRoleRequest) (*dir.DeleteUserRoleResponse, error) {

	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.DeleteUserRoleResponse{}, errors.Wrap(err, "failed to get directory")
	}

	return &dir.DeleteUserRoleResponse{}, eds.SetUserRole(tenantID, req.Id, NoAppl, req.Role, true)
}

func (s *DirectoryServer) GetUserPermissions(
	ctx context.Context,
	req *dir.GetUserPermissionsRequest) (*dir.GetUserPermissionsResponse, error) {

	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.GetUserPermissionsResponse{}, errors.Wrap(err, "failed to get directory")
	}

	permissions, err := eds.GetUserPermissions(tenantID, req.Id, NoAppl)
	if err != nil {
		return &dir.GetUserPermissionsResponse{}, errors.Wrapf(err, "get permissions")
	}

	return &dir.GetUserPermissionsResponse{
		Results: permissions,
	}, nil
}

func (s *DirectoryServer) SetUserPermissions(
	ctx context.Context,
	req *dir.SetUserPermissionsRequest) (*dir.SetUserPermissionsResponse, error) {

	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.SetUserPermissionsResponse{}, errors.Wrap(err, "failed to get directory")
	}

	return &dir.SetUserPermissionsResponse{}, eds.SetUserPermissions(tenantID, req.Id, NoAppl, req.Permissions, false)
}

func (s *DirectoryServer) SetUserPermission(
	ctx context.Context,
	req *dir.SetUserPermissionRequest) (*dir.SetUserPermissionResponse, error) {

	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.SetUserPermissionResponse{}, errors.Wrap(err, "failed to get directory")
	}

	return &dir.SetUserPermissionResponse{}, eds.SetUserPermission(tenantID, req.Id, NoAppl, req.Permission, false)
}

func (s *DirectoryServer) DeleteUserPermission(
	ctx context.Context,
	req *dir.DeleteUserPermissionRequest) (*dir.DeleteUserPermissionResponse, error) {

	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.DeleteUserPermissionResponse{}, errors.Wrap(err, "failed to get directory")
	}

	return &dir.DeleteUserPermissionResponse{}, eds.SetUserPermission(tenantID, req.Id, NoAppl, req.Permission, true)
}

func (s *DirectoryServer) ListUserApplications(
	ctx context.Context,
	req *dir.ListUserApplicationsRequest) (*dir.ListUserApplicationsResponse, error) {

	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.ListUserApplicationsResponse{}, errors.Wrap(err, "failed to get directory")
	}

	results, err := eds.ListUserApplications(tenantID, req.Id)
	if err != nil {
		return &dir.ListUserApplicationsResponse{}, errors.Wrapf(err, "list user application")
	}

	return &dir.ListUserApplicationsResponse{
		Results: results,
	}, nil
}

func (s *DirectoryServer) DeleteUserApplication(ctx context.Context, req *dir.DeleteUserApplicationRequest) (*dir.DeleteUserApplicationResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.DeleteUserApplicationResponse{}, errors.Wrap(err, "failed to get directory")
	}

	return &dir.DeleteUserApplicationResponse{}, eds.DeleteUserApplication(tenantID, req.Id, req.Name)
}

func (s *DirectoryServer) GetApplProperties(ctx context.Context, req *dir.GetApplPropertiesRequest) (*dir.GetApplPropertiesResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.GetApplPropertiesResponse{}, errors.Wrap(err, "failed to get directory")
	}

	properties, err := eds.GetUserProperties(tenantID, req.Id, req.Name)
	if err != nil {
		return &dir.GetApplPropertiesResponse{}, errors.Wrapf(err, "get properties")
	}

	return &dir.GetApplPropertiesResponse{
		Results: properties,
	}, nil
}

func (s *DirectoryServer) SetApplProperties(
	ctx context.Context,
	req *dir.SetApplPropertiesRequest) (*dir.SetApplPropertiesResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.SetApplPropertiesResponse{}, errors.Wrap(err, "failed to get directory")
	}

	return &dir.SetApplPropertiesResponse{}, eds.SetUserProperties(tenantID, req.Id, req.Name, req.Properties, false)
}

func (s *DirectoryServer) SetApplProperty(ctx context.Context, req *dir.SetApplPropertyRequest) (*dir.SetApplPropertyResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.SetApplPropertyResponse{}, errors.Wrap(err, "failed to get directory")
	}

	return &dir.SetApplPropertyResponse{}, eds.SetUserProperty(tenantID, req.Id, req.Name, req.Key, req.Value, false)
}

func (s *DirectoryServer) DeleteApplProperty(ctx context.Context, req *dir.DeleteApplPropertyRequest) (*dir.DeleteApplPropertyResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.DeleteApplPropertyResponse{}, errors.Wrap(err, "failed to get directory")
	}

	return &dir.DeleteApplPropertyResponse{}, eds.SetUserProperty(tenantID, req.Id, req.Name, req.Key, structpb.NewNullValue(), true)
}

func (s *DirectoryServer) GetApplRoles(ctx context.Context, req *dir.GetApplRolesRequest) (*dir.GetApplRolesResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.GetApplRolesResponse{}, errors.Wrap(err, "failed to get directory")
	}

	roles, err := eds.GetUserRoles(tenantID, req.Id, req.Name)
	if err != nil {
		if errors.Is(err, kvs.ErrKeyNotFound) {
			return &dir.GetApplRolesResponse{}, cerr.ErrUserNotFound.Err(err).Str("user-id", req.Id).Str("app-name", req.Name)
		}
		return &dir.GetApplRolesResponse{}, errors.Wrap(err, "get roles")
	}

	return &dir.GetApplRolesResponse{
		Results: roles,
	}, nil
}

func (s *DirectoryServer) SetApplRoles(ctx context.Context, req *dir.SetApplRolesRequest) (*dir.SetApplRolesResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.SetApplRolesResponse{}, errors.Wrap(err, "failed to get directory")
	}

	return &dir.SetApplRolesResponse{}, eds.SetUserRoles(tenantID, req.Id, req.Name, req.Roles, false)
}

func (s *DirectoryServer) SetApplRole(ctx context.Context, req *dir.SetApplRoleRequest) (*dir.SetApplRoleResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.SetApplRoleResponse{}, errors.Wrap(err, "failed to get directory")
	}

	return &dir.SetApplRoleResponse{}, eds.SetUserRole(tenantID, req.Id, req.Name, req.Role, false)
}

func (s *DirectoryServer) DeleteApplRole(ctx context.Context, req *dir.DeleteApplRoleRequest) (*dir.DeleteApplRoleResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.DeleteApplRoleResponse{}, errors.Wrap(err, "failed to get directory")
	}

	return &dir.DeleteApplRoleResponse{}, eds.SetUserRole(tenantID, req.Id, req.Name, req.Role, true)
}

func (s *DirectoryServer) GetApplPermissions(ctx context.Context, req *dir.GetApplPermissionsRequest) (*dir.GetApplPermissionsResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.GetApplPermissionsResponse{}, errors.Wrap(err, "failed to get directory")
	}

	permissions, err := eds.GetUserPermissions(tenantID, req.Id, req.Name)
	if err != nil {
		return &dir.GetApplPermissionsResponse{}, errors.Wrapf(err, "get permissions")
	}

	return &dir.GetApplPermissionsResponse{
		Results: permissions,
	}, nil
}

func (s *DirectoryServer) SetApplPermissions(ctx context.Context, req *dir.SetApplPermissionsRequest) (*dir.SetApplPermissionsResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.SetApplPermissionsResponse{}, errors.Wrap(err, "failed to get directory")
	}

	return &dir.SetApplPermissionsResponse{}, eds.SetUserPermissions(tenantID, req.Id, req.Name, req.Permissions, false)
}

func (s *DirectoryServer) SetApplPermission(ctx context.Context, req *dir.SetApplPermissionRequest) (*dir.SetApplPermissionResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.SetApplPermissionResponse{}, errors.Wrap(err, "failed to get directory")
	}

	return &dir.SetApplPermissionResponse{}, eds.SetUserPermission(tenantID, req.Id, req.Name, req.Permission, false)
}

func (s *DirectoryServer) DeleteApplPermission(ctx context.Context, req *dir.DeleteApplPermissionRequest) (*dir.DeleteApplPermissionResponse, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	eds, err := s.eds(ctx, "")
	if err != nil {
		return &dir.DeleteApplPermissionResponse{}, errors.Wrap(err, "failed to get directory")
	}

	return &dir.DeleteApplPermissionResponse{}, eds.SetUserPermission(tenantID, req.Id, req.Name, req.Permission, true)
}
