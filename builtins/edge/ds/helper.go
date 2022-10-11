package ds

import (
	"bytes"
	"io"

	v2 "github.com/aserto-dev/go-directory/aserto/directory/common/v2"
	"github.com/google/uuid"
	"github.com/open-policy-agent/opa/ast"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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

func ValidateObject(obj *v2.ObjectIdentifier) bool {
	if obj != nil && obj.Id != nil && *obj.Id != "" {
		return true
	}

	if obj != nil && obj.Type != nil && *obj.Type != "" && obj.Key != nil && *obj.Key != "" {
		return true
	}
	return false
}

func ValidateRelationType(rel *v2.RelationTypeIdentifier) bool {
	if rel != nil && rel.Id != nil {
		return true
	}

	if rel != nil && rel.Name != nil && *rel.Name != "" && rel.ObjectType != nil && *rel.ObjectType != "" {
		return true
	}
	return false
}

func ValidatePermissionType(perm *v2.PermissionIdentifier) bool {
	if perm != nil && perm.Id != nil && *perm.Id != "" {
		return true
	}
	if perm != nil && perm.Name != nil && *perm.Name != "" {
		return true
	}
	return false
}

func ValidateRelation(relation *v2.RelationIdentifier) bool {
	if ValidateObject(relation.Object) && ValidateRelationType(relation.Relation) {
		return true
	}
	return false
}

func StrPrt(input string) *string {
	return &input
}

// ProtoToBuf, marshal proto message to buffer.
func ProtoToBuf(w io.Writer, msg proto.Message) error {
	b, err := protojson.MarshalOptions{
		Multiline:       false,
		Indent:          "",
		AllowPartial:    false,
		UseProtoNames:   true,
		UseEnumNumbers:  false,
		EmitUnpopulated: false,
	}.Marshal(msg)
	if err != nil {
		return err
	}

	if _, err := w.Write(b); err != nil {
		return err
	}

	return nil
}

// BufToProto, unmarshal buffer to proto message instance.
func BufToProto(r io.Reader, msg proto.Message) error {
	buf := new(bytes.Buffer)

	if _, err := buf.ReadFrom(r); err != nil {
		return err
	}

	return protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}.Unmarshal(buf.Bytes(), msg)
}
