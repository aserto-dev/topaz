package ds

import (
	v2 "github.com/aserto-dev/go-directory/aserto/directory/common/v2"
	"github.com/google/uuid"
	"github.com/open-policy-agent/opa/ast"
)

func IsValidID(id string) bool {
	if len(id) != 36 {
		return false
	}
	_, err := uuid.Parse(id)
	return err == nil
}

func help(fnName string, args interface{}) (*ast.Term, error) {
	m := map[string]interface{}{fnName: args}
	val, err := ast.InterfaceToValue(m)
	if err != nil {
		return nil, err
	}
	return ast.NewTerm(val), nil
}

func ValidateObject(o *v2.ObjectIdentifier) bool {
	if o != nil && o.Id != nil && *o.Id != "" {
		return true
	}

	if o != nil && o.Type != nil && *o.Type != "" && o.Key != nil && *o.Key != "" {
		return true
	}
	return false
}

func ValidateRelationType(o *v2.RelationTypeIdentifier) bool {
	if o != nil && o.Id != nil {
		return true
	}

	if o != nil && o.Name != nil && *o.Name != "" && o.ObjectType != nil && *o.ObjectType != "" {
		return true
	}
	return false
}

func ValidatePermissionType(o *v2.PermissionIdentifier) bool {
	if o != nil && o.Id != nil && *o.Id != "" {
		return true
	}
	if o != nil && o.Name != nil && *o.Name != "" {
		return true
	}
	return false
}

func ValidateRelation(o *v2.RelationIdentifier) bool {
	if ValidateObject(o.Object) && ValidateRelationType(o.Relation) {
		return true
	}
	return false
}

func StrPrt(input string) *string {
	return &input
}
