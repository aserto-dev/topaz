package parser_test

import (
	"testing"

	"github.com/aserto-dev/topaz/azm/model"
	"github.com/aserto-dev/topaz/azm/parser"
	"github.com/stretchr/testify/require"
)

func TestRelationParser(t *testing.T) {
	for _, test := range relationTests {
		t.Run(test.input, func(tt *testing.T) {
			rel, err := parser.ParseRelation(test.input)
			require.NoError(t, err)
			test.validate(rel, require.New(tt))
		})
	}
}

func TestPermissionParser(t *testing.T) {
	for _, test := range permissionTests {
		t.Run(test.input, func(tt *testing.T) {
			perm, err := parser.ParsePermission(test.input)
			require.NoError(t, err)
			test.validate(perm, require.New(tt))
		})
	}
}

type relationTest struct {
	input    string
	validate func([]*model.RelationRef, *require.Assertions)
}

type permissionTest struct {
	input    string
	validate func(*model.Permission, *require.Assertions)
}

var relationTests = []relationTest{
	{
		"user",
		func(rel []*model.RelationRef, assert *require.Assertions) {
			assert.Len(rel, 1)
			term := rel[0]
			assert.True(term.IsDirect())
			assert.Equal(model.ObjectName("user"), term.Object)
			assert.Empty(term.Relation)
		},
	},
	{
		"name-with-dashes",
		func(rel []*model.RelationRef, assert *require.Assertions) {
			assert.Len(rel, 1)
			term := rel[0]
			assert.True(term.IsDirect())
			assert.Equal(model.ObjectName("name-with-dashes"), term.Object)
			assert.Empty(term.Relation)
		},
	},
	{
		"group#member",
		func(rel []*model.RelationRef, assert *require.Assertions) {
			assert.Len(rel, 1)
			term := rel[0]
			assert.True(term.IsSubject())
			assert.Equal(model.ObjectName("group"), term.Object)
			assert.Equal(model.RelationName("member"), term.Relation)
		},
	},
	{
		"user:*",
		func(rel []*model.RelationRef, assert *require.Assertions) {
			assert.Len(rel, 1)
			term := rel[0]
			assert.True(term.IsWildcard())
			assert.Equal(model.ObjectName("user"), term.Object)
			assert.Equal(model.RelationName("*"), term.Relation)
		},
	},
	{
		"user | group",
		func(rel []*model.RelationRef, assert *require.Assertions) {
			assert.Len(rel, 2)

			assert.True(rel[0].IsDirect())
			assert.Equal(model.ObjectName("user"), rel[0].Object)
			assert.Empty(rel[0].Relation)

			assert.True(rel[1].IsDirect())
			assert.Equal(model.ObjectName("group"), rel[1].Object)
			assert.Empty(rel[1].Relation)
		},
	},
	{
		"user | group | user:* | group#member",
		func(rel []*model.RelationRef, assert *require.Assertions) {
			assert.Len(rel, 4)

			assert.True(rel[0].IsDirect())
			assert.Equal(model.ObjectName("user"), rel[0].Object)
			assert.Empty(rel[0].Relation)

			assert.True(rel[0].IsDirect())
			assert.Equal(model.ObjectName("group"), rel[1].Object)
			assert.Empty(rel[1].Relation)

			assert.True(rel[2].IsWildcard())
			assert.Equal(model.ObjectName("user"), rel[2].Object)

			assert.True(rel[3].IsSubject())
			assert.Equal(model.ObjectName("group"), rel[3].Object)
			assert.Equal(model.RelationName("member"), rel[3].Relation)
		},
	},
	{
		"IDENTITY",
		func(rel []*model.RelationRef, assert *require.Assertions) {
			assert.Len(rel, 1)
			term := rel[0]
			assert.True(term.IsDirect())
			assert.Equal(model.ObjectName("IDENTITY"), term.Object)
			assert.Empty(term.Relation)
		},
	},
}

var permissionTests = []permissionTest{
	{
		"can_write",
		func(perm *model.Permission, assert *require.Assertions) {
			assert.Equal(model.RelationName("can_write"), perm.Union[0].RelOrPerm)
			assert.Empty(perm.Union[0].Base)
			assert.Empty(perm.Intersection)
			assert.Nil(perm.Exclusion)
		},
	},
	{
		"can_write | parent->can_read",
		func(perm *model.Permission, assert *require.Assertions) {
			assert.Len(perm.Union, 2)
			assert.Equal(model.RelationName("can_write"), perm.Union[0].RelOrPerm)
			assert.Empty(perm.Union[0].Base)
			assert.Equal(model.RelationName("parent"), perm.Union[1].Base)
			assert.Equal(model.RelationName("can_read"), perm.Union[1].RelOrPerm)
		},
	},
	{
		"can_write & can_read",
		func(perm *model.Permission, assert *require.Assertions) {
			assert.Len(perm.Intersection, 2)
			assert.Equal(model.RelationName("can_write"), perm.Intersection[0].RelOrPerm)
			assert.Empty(perm.Intersection[0].Base)
			assert.Equal(model.RelationName("can_read"), perm.Intersection[1].RelOrPerm)
			assert.Empty(perm.Intersection[1].Base)
		},
	},
	{
		"can_write - can_read",
		func(perm *model.Permission, assert *require.Assertions) {
			assert.Equal(model.RelationName("can_write"), perm.Exclusion.Include.RelOrPerm)
			assert.Empty(perm.Exclusion.Include.Base)
			assert.Equal(model.RelationName("can_read"), perm.Exclusion.Exclude.RelOrPerm)
			assert.Empty(perm.Exclusion.Exclude.Base)
		},
	},
	{
		"parent->can_read - parent->can_write",
		func(perm *model.Permission, assert *require.Assertions) {
			assert.Equal(model.RelationName("parent"), perm.Exclusion.Include.Base)
			assert.Equal(model.RelationName("can_read"), perm.Exclusion.Include.RelOrPerm)
			assert.Equal(model.RelationName("parent"), perm.Exclusion.Exclude.Base)
			assert.Equal(model.RelationName("can_write"), perm.Exclusion.Exclude.RelOrPerm)
		},
	},
}
