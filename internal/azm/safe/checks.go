package safe

import (
	"iter"

	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
)

type SafeChecks struct {
	*dsr3.ChecksRequest
}

func Checks(i *dsr3.ChecksRequest) *SafeChecks {
	if i.GetDefault() == nil {
		i.Default = &dsr3.CheckRequest{}
	}

	return &SafeChecks{i}
}

// CheckRequests returns an iterator that materializes all checks in order.
func (c *SafeChecks) CheckRequests() iter.Seq[SafeCheck] {
	return func(yield func(SafeCheck) bool) {
		defaults := &dsc3.RelationIdentifier{
			ObjectType:  c.GetDefault().GetObjectType(),
			ObjectId:    c.GetDefault().GetObjectId(),
			Relation:    c.GetDefault().GetRelation(),
			SubjectType: c.GetDefault().GetSubjectType(),
			SubjectId:   c.GetDefault().GetSubjectId(),
		}

		for _, check := range c.Checks {
			req := &dsr3.CheckRequest{
				ObjectType:  fallback(check.GetObjectType(), defaults.GetObjectType()),
				ObjectId:    fallback(check.GetObjectId(), defaults.GetObjectId()),
				Relation:    fallback(check.GetRelation(), defaults.GetRelation()),
				SubjectType: fallback(check.GetSubjectType(), defaults.GetSubjectType()),
				SubjectId:   fallback(check.GetSubjectId(), defaults.GetSubjectId()),
			}
			if !yield(SafeCheck{req}) {
				break
			}
		}
	}
}

func fallback[T comparable](val, fallback T) T {
	var def T
	if val == def {
		return fallback
	}

	return val
}
