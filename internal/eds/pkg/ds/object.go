package ds

import (
	"bytes"

	dsc3 "github.com/aserto-dev/topaz/api/directory/v4"
	"github.com/aserto-dev/topaz/azm/safe"
	"github.com/aserto-dev/topaz/internal/eds/pkg/x"
)

type object struct {
	*safe.SafeObject
}

func Object(i *dsc3.Object) *object { return &object{safe.Object(i)} }

func (i *object) StrKey() string {
	return i.GetObjectType() + string(TypeIDSeparator) + i.GetObjectId()
}

func (i *object) Key() []byte {
	var buf bytes.Buffer

	buf.Grow(x.MaxObjectIdentifierSize)

	buf.WriteString(i.GetObjectType())
	buf.WriteByte(TypeIDSeparator)
	buf.WriteString(i.GetObjectId())

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

const (
	propDisplayName string = "display_name"
)

// PatchObjectRead: transfers Objects.Properties["display_name"] to Object.DisplayName (BACKWARD COMPATIBILITY).
func PatchObjectRead(i *dsc3.Object) *dsc3.Object {
	// fields := i.GetProperties().GetFields()

	// if displayName, ok := fields[propDisplayName]; ok {
	// 	i.DisplayName = displayName.GetStringValue() //nolint:staticcheck // Marked as deprecated
	// }

	return i
}

// PatchObjectWrite: transfers Object.DisplayName to Objects.Properties["display_name"] (FORWARD COMPATIBILITY).
func PatchObjectWrite(i *dsc3.Object) *dsc3.Object {
	// if displayName := i.GetDisplayName(); len(displayName) > 0 { //nolint:staticcheck // Marked as deprecated
	// 	if i.GetProperties().Fields == nil {
	// 		i.Properties.Fields = map[string]*structpb.Value{}
	// 	}

	// 	i.Properties.Fields[propDisplayName] = structpb.NewStringValue(displayName)
	// 	i.DisplayName = "" //nolint:staticcheck // Marked as deprecated
	// }

	return i
}
