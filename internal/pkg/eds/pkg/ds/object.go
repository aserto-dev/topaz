package ds

import (
	"bytes"

	"github.com/aserto-dev/azm/safe"
	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/x"
)

type object struct {
	*safe.SafeObject
}

func Object(i *dsc3.Object) *object { return &object{safe.Object(i)} }

func (i *object) StrKey() string {
	return i.GetType() + string(TypeIDSeparator) + i.GetId()
}

func (i *object) Key() []byte {
	var buf bytes.Buffer

	buf.Grow(x.MaxObjectIdentifierSize)

	buf.WriteString(i.GetType())
	buf.WriteByte(TypeIDSeparator)
	buf.WriteString(i.GetId())

	return buf.Bytes()
}

type objectIdentifier struct {
	*safe.SafeObjectIdentifier
}

func ObjectIdentifier(i *dsc3.ObjectIdentifier) *objectIdentifier {
	return &objectIdentifier{safe.ObjectIdentifier(i)}
}

func (i *objectIdentifier) StrKey() string {
	return i.GetObjectType() + string(TypeIDSeparator) + i.GetObjectId()
}

func (i *objectIdentifier) Key() []byte {
	var buf bytes.Buffer

	buf.Grow(x.MaxObjectIdentifierSize)

	buf.WriteString(i.GetObjectType())
	buf.WriteByte(TypeIDSeparator)

	if i.GetObjectId() != "" {
		buf.WriteString(i.GetObjectId())
	}

	return buf.Bytes()
}

func ObjectSelector(i *dsc3.ObjectIdentifier) *safe.SafeObjectSelector {
	return safe.ObjectSelector(i)
}
