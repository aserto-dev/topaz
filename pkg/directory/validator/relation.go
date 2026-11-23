package validator

import (
	dsc3 "github.com/aserto-dev/topaz/api/directory/v3/common"
	dsr3 "github.com/aserto-dev/topaz/api/directory/v3/reader"
	dsw3 "github.com/aserto-dev/topaz/api/directory/v3/writer"
)

func Relation(msg *dsc3.Relation) error {
	if msg == nil {
		return nil
	}

	if err := TypeIdentifier(fieldObjectType, msg.GetObjectType()); err != nil {
		return err
	}

	if err := InstanceIdentifier(fieldObjectID, msg.GetObjectId()); err != nil {
		return err
	}

	if err := TypeIdentifier(fieldRelation, msg.GetRelation()); err != nil {
		return err
	}

	if err := TypeIdentifier(fieldSubjectType, msg.GetSubjectType()); err != nil {
		return err
	}

	if err := InstanceIdentifier(fieldSubjectID, msg.GetSubjectId()); err != nil {
		return err
	}

	if err := TypeIdentifier(fieldSubjectRelation, msg.GetSubjectRelation()); err != nil {
		return err
	}

	if err := Etag(fieldETag, msg.GetEtag()); err != nil {
		return err
	}

	return nil
}

func RelationIdentifier(msg *dsc3.RelationIdentifier) error {
	if msg == nil {
		return nil
	}

	if err := TypeIdentifier(fieldObjectType, msg.GetObjectType()); err != nil {
		return err
	}

	if err := InstanceIdentifier(fieldObjectID, msg.GetObjectId()); err != nil {
		return err
	}

	if err := TypeIdentifier(fieldRelation, msg.GetRelation()); err != nil {
		return err
	}

	if err := TypeIdentifier(fieldSubjectType, msg.GetSubjectType()); err != nil {
		return err
	}

	if err := InstanceIdentifier(fieldSubjectID, msg.GetSubjectId()); err != nil {
		return err
	}

	if err := TypeIdentifier(fieldSubjectRelation, msg.GetSubjectRelation()); err != nil {
		return err
	}

	return nil
}

func GetRelationRequest(msg *dsr3.GetRelationRequest) error {
	if msg == nil {
		return nil
	}

	if err := TypeIdentifier(fieldObjectType, msg.GetObjectType()); err != nil {
		return err
	}

	if err := InstanceIdentifier(fieldObjectID, msg.GetObjectId()); err != nil {
		return err
	}

	if err := TypeIdentifier(fieldRelation, msg.GetRelation()); err != nil {
		return err
	}

	if err := TypeIdentifier(fieldSubjectType, msg.GetSubjectType()); err != nil {
		return err
	}

	if err := InstanceIdentifier(fieldSubjectID, msg.GetSubjectId()); err != nil {
		return err
	}

	if err := TypeIdentifier(fieldSubjectRelation, msg.GetSubjectRelation()); err != nil {
		return err
	}

	if err := IdentifierTypePresence(fieldObjectID, fieldObjectType, msg.GetObjectId(), msg.GetObjectType()); err != nil {
		return err
	}

	if err := IdentifierTypePresence(fieldSubjectID, fieldSubjectType, msg.GetSubjectId(), msg.GetSubjectType()); err != nil {
		return err
	}

	return nil
}

func GetRelationsRequest(msg *dsr3.GetRelationsRequest) error {
	if msg == nil {
		return nil
	}

	if err := TypeIdentifier(fieldObjectType, msg.GetObjectType()); err != nil {
		return err
	}

	if err := InstanceIdentifier(fieldObjectID, msg.GetObjectId()); err != nil {
		return err
	}

	if err := TypeIdentifier(fieldRelation, msg.GetRelation()); err != nil {
		return err
	}

	if err := TypeIdentifier(fieldSubjectType, msg.GetSubjectType()); err != nil {
		return err
	}

	if err := InstanceIdentifier(fieldSubjectID, msg.GetSubjectId()); err != nil {
		return err
	}

	if err := TypeIdentifier(fieldSubjectRelation, msg.GetSubjectRelation()); err != nil {
		return err
	}

	if err := IdentifierTypePresence(fieldObjectID, fieldObjectType, msg.GetObjectId(), msg.GetObjectType()); err != nil {
		return err
	}

	if err := IdentifierTypePresence(fieldSubjectID, fieldSubjectType, msg.GetSubjectId(), msg.GetSubjectType()); err != nil {
		return err
	}

	if err := PaginationRequest(msg.GetPage()); err != nil {
		return err
	}

	return nil
}

func SetRelationRequest(msg *dsw3.SetRelationRequest) error {
	if msg == nil || msg.GetRelation() == nil {
		return nil
	}

	if err := Relation(msg.GetRelation()); err != nil {
		return err
	}

	return nil
}

func DeleteRelationRequest(msg *dsw3.DeleteRelationRequest) error {
	if msg == nil {
		return nil
	}

	if err := TypeIdentifier(fieldObjectType, msg.GetObjectType()); err != nil {
		return err
	}

	if err := InstanceIdentifier(fieldObjectID, msg.GetObjectId()); err != nil {
		return err
	}

	if err := TypeIdentifier(fieldRelation, msg.GetRelation()); err != nil {
		return err
	}

	if err := TypeIdentifier(fieldSubjectType, msg.GetSubjectType()); err != nil {
		return err
	}

	if err := InstanceIdentifier(fieldSubjectID, msg.GetSubjectId()); err != nil {
		return err
	}

	if err := TypeIdentifier(fieldSubjectRelation, msg.GetSubjectRelation()); err != nil {
		return err
	}

	return nil
}
