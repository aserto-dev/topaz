package graph

import (
	"fmt"

	dsc "github.com/aserto-dev/topaz/api/directory/v4"
	"github.com/aserto-dev/topaz/azm/model"
)

// relation is a comparable representation of a relation. It can be used as a map key.
type relation struct {
	ot   model.ObjectName
	oid  ObjectID
	rel  model.RelationName
	st   model.ObjectName
	sid  ObjectID
	srel model.RelationName

	tail model.RelationName
}

type relations []*relation

// converts a dsc.RelationIdentifier to a relation.
func relationFromProto(rel *dsc.RelationIdentifier) *relation {
	return &relation{
		ot:   model.ObjectName(rel.GetObjectType()),
		oid:  ObjectID(rel.GetObjectId()),
		rel:  model.RelationName(rel.GetRelation()),
		st:   model.ObjectName(rel.GetSubjectType()),
		sid:  ObjectID(rel.GetSubjectId()),
		srel: model.RelationName(rel.GetSubjectRelation()),
	}
}

func (r *relation) String() string {
	str := fmt.Sprintf("%s:%s#%s@%s:%s", r.ot, displayID(r.oid), r.rel, r.st, displayID(r.sid))
	if r.srel != "" {
		str += fmt.Sprintf("#%s", r.srel)
	}

	return str
}

func (r *relation) asProto() *dsc.RelationIdentifier {
	return &dsc.RelationIdentifier{
		ObjectType:      string(r.ot),
		ObjectId:        string(r.oid),
		Relation:        string(r.rel),
		SubjectType:     string(r.st),
		SubjectId:       string(r.sid),
		SubjectRelation: string(r.srel),
	}
}

func (r *relation) subject() *object {
	return &object{
		Type: r.st,
		ID:   r.sid,
	}
}

func displayID(id ObjectID) string {
	if id == "" {
		return "?"
	}

	return id.String()
}
