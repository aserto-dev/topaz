package validator

import (
	dsc "github.com/aserto-dev/topaz/api/directory/v4"
	dsr "github.com/aserto-dev/topaz/api/directory/v4/reader"
	dsw "github.com/aserto-dev/topaz/api/directory/v4/writer"
)

func Relation(msg *dsc.Relation) error {
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

func RelationIdentifier(msg *dsc.RelationIdentifier) error {
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

func GetRelationRequest(msg *dsr.GetRelationRequest) error {
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

func ListRelationsRequest(msg *dsr.ListRelationsRequest) error {
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

func SetRelationRequest(msg *dsw.SetRelationRequest) error {
	if msg == nil || msg.GetRelation() == nil {
		return nil
	}

	if err := Relation(msg.GetRelation()); err != nil {
		return err
	}

	return nil
}

func DeleteRelationRequest(msg *dsw.DeleteRelationRequest) error {
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
