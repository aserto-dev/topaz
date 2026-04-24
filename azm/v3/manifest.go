package v3

import (
	"strconv"

	"github.com/aserto-dev/topaz/api/directory/pkg/derr"
	"github.com/aserto-dev/topaz/azm"
	"github.com/hashicorp/go-multierror"
	"gopkg.in/yaml.v3"
)

const SupportedSchemaVersion int = 3

type Manifest struct {
	ModelInfo   *ModelInfo                     `yaml:"model"`
	ObjectTypes map[ObjectTypeName]*ObjectType `yaml:"types"`
}

type SchemaVersion int

type ModelInfo struct {
	Version SchemaVersion `yaml:"version"`
}

type (
	ObjectTypeName string
	RelationName   string
	PermissionName string
)

type ObjectType struct {
	Relations   map[RelationName]string   `yaml:"relations,omitempty"`
	Permissions map[PermissionName]string `yaml:"permissions,omitempty"`
}

func (m *Manifest) ValidateNames() error {
	var errs error

	for on, o := range m.ObjectTypes {
		if !ValidIdentifier(on) {
			errs = multierror.Append(errs, derr.ErrInvalidObjectType.Msgf("invalid name '%s'", on))
		}

		if o == nil {
			continue
		}

		for rn := range o.Relations {
			if !ValidIdentifier(rn) {
				errs = multierror.Append(errs, derr.ErrInvalidRelationType.Msgf("invalid name '%s:%s'", on, rn))
			}
		}

		for pn := range o.Permissions {
			if !ValidIdentifier(pn) {
				errs = multierror.Append(errs, derr.ErrInvalidPermission.Msgf("invalid name '%s:%s'", on, pn))
			}
		}
	}

	return errs
}

func (v *SchemaVersion) UnmarshalYAML(value *yaml.Node) error {
	version, err := strconv.Atoi(value.Value)
	if err != nil {
		return err
	}

	if version != SupportedSchemaVersion {
		return azm.ErrInvalidSchemaVersion.Msgf("%d", version)
	}

	*v = SchemaVersion(version)

	return nil
}
