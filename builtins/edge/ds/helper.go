package ds

import (
	ds2 "github.com/aserto-dev/go-directory/aserto/directory/v2"
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

type ObjectParam struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Key  string `json:"key"`
}

func (o *ObjectParam) Validate() *ds2.ObjectParam {
	if o != nil && o.ID != "" {
		return &ds2.ObjectParam{
			Opt: &ds2.ObjectParam_Id{
				Id: o.ID,
			},
		}
	}

	if o != nil && o.Type != "" && o.Key != "" {
		return &ds2.ObjectParam{
			Opt: &ds2.ObjectParam_Key{
				Key: &ds2.ObjectKey{
					Type: o.Type,
					Key:  o.Key,
				},
			},
		}
	}
	return nil
}

type RelationTypeParam struct {
	ObjectType string `json:"object_type"`
	Name       string `json:"name"`
}

func (r *RelationTypeParam) Validate() *ds2.RelationTypeParam {
	if r != nil && r.ObjectType != "" && r.Name != "" {
		return &ds2.RelationTypeParam{
			Opt: &ds2.RelationTypeParam_Key{
				Key: &ds2.RelationTypeKey{
					ObjectType: r.ObjectType,
					Name:       r.Name,
				},
			},
		}
	}
	return nil
}

type PermissionTypeParam struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (p *PermissionTypeParam) Validate() *ds2.PermissionParam {
	if p != nil && p.ID != "" {
		return &ds2.PermissionParam{
			Opt: &ds2.PermissionParam_Id{
				Id: p.ID,
			},
		}
	}
	if p != nil && p.Name != "" {
		return &ds2.PermissionParam{
			Opt: &ds2.PermissionParam_Name{
				Name: p.Name,
			},
		}
	}
	return nil
}

type RelationParam struct {
	SubjectType string `json:"subject_type"`
	SubjectID   string `json:"subject_id"`
	Relation    string `json:"relation"`
	ObjectType  string `json:"object_type"`
	ObjectID    string `json:"object_id"`
}

func (r *RelationParam) Validate() *ds2.RelationParam {
	if r == nil {
		return nil
	}
	if r.ObjectID != "" || r.ObjectType != "" || r.Relation != "" || r.SubjectID != "" || r.SubjectType != "" {
		return &ds2.RelationParam{}
	}
	return nil
}
