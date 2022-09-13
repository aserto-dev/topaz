package directory

import (
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	info "github.com/aserto-dev/go-grpc/aserto/common/info/v1"
	op "github.com/aserto-dev/go-utils/opts"
	"google.golang.org/protobuf/types/known/structpb"
)

type Directory interface {
	GetUserFromIdentity(tenantID, ident string) (user *api.User, err error)
	GetIdentity(tenantID, ident string) (string, error)
	GetUser(tenantID, uid string) (*api.User, error)
	GetUserExt(tenantID, uid string) (*api.User, error)
	ListUsers(tenantID, pageToken string, pageSize int32, paths []string, params ...op.Param) ([]*api.User, string, int32, error)
	ListUsersExt(tenantID, pageToken string, pageSize int32, paths []string, params ...op.Param) ([]*api.User, string, int32, error)
	CreateUser(tenantID string, user *api.User) (*api.User, error)
	UpdateUser(tenantID string, user *api.User) (*api.User, error)
	UpsertUser(tenantID string, user *api.User) (*api.User, bool, error)
	UserExists(tenantID, uid string) bool
	DeleteUser(tenantID, uid string) error
	ListTenants() ([]string, error)
	TenantExists(tenantID string) bool
	CreateTenant(tenantID string) error
	DeleteTenant(tenantID string) error
	GetUserProperties(tenantID, uid, app string) (*structpb.Struct, error)
	SetUserProperties(tenantID, uid, app string, properties *structpb.Struct, remove bool) error
	SetUserProperty(tenantID, uid, app, key string, value *structpb.Value, remove bool) error
	GetUserRoles(tenantID, uid, app string) ([]string, error)
	SetUserRoles(tenantID, uid, app string, roles []string, remove bool) error
	SetUserRole(tenantID, uid, app, role string, remove bool) error
	GetUserPermissions(tenantID, uid, app string) ([]string, error)
	SetUserPermissions(tenantID, uid, app string, permissions []string, remove bool) error
	SetUserPermission(tenantID, uid, app, permission string, remove bool) error
	ListUserApplications(tenantID, uid string) (applications []string, err error)
	DeleteUserApplication(tenantID, uid, name string) error
	UpdateUserExt(tenantID string, ext *api.UserExt) error
	ListResources(tenantID, pageToken string, pageSize int32) ([]string, string, int32, error)
	GetResource(tenantID, key string) (*structpb.Struct, error)
	SetResource(tenantID, key string, value *structpb.Struct) error
	DeleteResource(tenantID, key string) error
	GetVersionInfo() (*info.VersionInfo, error)
	GetSystemInfo() (*info.SystemInfo, error)
	GetValue(path []string, key string) (*structpb.Struct, error)
}
