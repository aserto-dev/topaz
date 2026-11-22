package validator

type Field string

const (
	fieldObjectType      = Field("object_type")
	fieldObjectID        = Field("object_id")
	fieldRelation        = Field("relation")
	fieldSubjectType     = Field("subject_type")
	fieldSubjectID       = Field("subject_id")
	fieldSubjectRelation = Field("subject_relation")
	fieldDisplayName     = Field("display_name")
	fieldETag            = Field("etag")
	fieldType            = Field("type")
	fieldID              = Field("id")
)
