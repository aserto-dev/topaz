package validator

import (
	dsr3 "github.com/aserto-dev/topaz/api/directory/v3/reader"
)

func CheckRequest(msg *dsr3.CheckRequest) error {
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

	return nil
}

func ChecksRequest(msg *dsr3.ChecksRequest) error {
	if msg == nil {
		return nil
	}

	if msg.GetDefault() != nil {
		if err := CheckRequest(msg.GetDefault()); err != nil {
			return err
		}
	}

	for _, check := range msg.GetChecks() {
		if err := CheckRequest(check); err != nil {
			return err
		}
	}

	return nil
}

// CheckPermissionRequest gets handled by CheckRequest, hence returns nil.
func CheckPermissionRequest(msg *dsr3.CheckPermissionRequest) error {
	return nil
}

// CheckRelationRequest gets handled by CheckRequest, hence returns nil.
func CheckRelationRequest(msg *dsr3.CheckRelationRequest) error {
	return nil
}

func GetGraphRequest(msg *dsr3.GetGraphRequest) error {
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
