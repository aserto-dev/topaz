package v3

import (
	"io"
	"os"

	"github.com/aserto-dev/topaz/api/directory/pkg/derr"
	"github.com/aserto-dev/topaz/azm/model"
	"github.com/aserto-dev/topaz/azm/parser"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

func Load(r io.Reader) (*model.Model, error) {
	m := model.Model{
		Version: model.ModelVersion,
		Objects: map[model.ObjectName]*model.Object{},
	}

	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)

	manifest := Manifest{}
	if err := dec.Decode(&manifest); err != nil {
		if errors.Is(err, io.EOF) {
			return &m, nil
		}

		return nil, err
	}

	if err := manifest.ValidateNames(); err != nil {
		return nil, derr.ErrInvalidArgument.Err(err)
	}

	var errs error

	for on, o := range manifest.ObjectTypes {
		log.Debug().Str("object", string(on)).Msg("loading object")

		if o == nil {
			o = &ObjectType{}
		}

		relations, errors := parseRelations(on, o)
		if len(errors) > 0 {
			errs = multierror.Append(errs, errors...)
		}

		permissions, errors := parsePermissions(on, o, relations)
		if len(errors) > 0 {
			errs = multierror.Append(errs, errors...)
		}

		m.Objects[model.ObjectName(on)] = &model.Object{
			Relations:   relations,
			Permissions: permissions,
		}
	}

	if errs != nil {
		return &m, derr.ErrInvalidArgument.Err(errs)
	}

	return &m, m.Validate()
}

func LoadFile(path string) (*model.Model, error) {
	f, err := os.Open(path) //nolint:gosec  // this is intended to parse a model from a user-provided path.
	if err != nil {
		return nil, err
	}

	defer func() { _ = f.Close() }()

	return Load(f)
}

func parseRelations(on ObjectTypeName, o *ObjectType) (model.Relations, []error) {
	var errs []error

	relationTerms := lo.MapEntries(o.Relations, func(name RelationName, rd string) (model.RelationName, []*model.RelationRef) {
		log.Debug().Str("object", string(on)).Str("relation", string(name)).Msg("loading relation")
		rn := model.RelationName(name)

		if rd == "" {
			errs = append(errs, derr.ErrInvalidRelationType.Msgf("relation '%s:%s' has empty definition", on, rn))
			return rn, nil
		}

		rel, err := parser.ParseRelation(rd)
		if err != nil {
			errs = append(errs, derr.ErrInvalidRelationType.Err(err).Msgf("'%s:%s'", on, rn))
		}

		return rn, rel
	})

	relations := lo.MapEntries(relationTerms, func(rn model.RelationName, rts []*model.RelationRef) (model.RelationName, *model.Relation) {
		return rn, &model.Relation{Union: rts}
	})

	return relations, errs
}

func parsePermissions(on ObjectTypeName, o *ObjectType, relations model.Relations) (model.Permissions, []error) {
	var errs []error

	permissions := lo.MapEntries(o.Permissions, func(name PermissionName, pd string) (model.RelationName, *model.Permission) {
		log.Debug().Str("object", string(on)).Str("permission", string(name)).Msg("loading permission")
		pn := model.RelationName(name)

		if _, ok := relations[pn]; ok {
			errs = append(errs, derr.ErrInvalidPermission.Msgf(
				"permission name '%[1]s:%[2]s' conflicts with relation '%[1]s:%[2]s'", on, pn),
			)
		}

		if pd == "" {
			errs = append(errs, derr.ErrInvalidPermission.Msgf("permission '%s:%s' has empty definition", on, pn))
			return pn, nil
		}

		perm, err := parser.ParsePermission(pd)
		if err != nil {
			errs = append(errs, derr.ErrInvalidPermission.Err(err).Msgf("%s:%s", on, pn))
		}

		return pn, perm
	})

	return permissions, errs
}
