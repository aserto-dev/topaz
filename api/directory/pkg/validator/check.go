package validator

import (
	dsr "github.com/aserto-dev/topaz/api/directory/v4/reader"
)

func CheckRequest(msg *dsr.CheckRequest) error {
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

func ChecksRequest(msg *dsr.ChecksRequest) error {
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
// func CheckPermissionRequest(msg *dsr.CheckPermissionRequest) error {
// 	return nil
// }

// CheckRelationRequest gets handled by CheckRequest, hence returns nil.
// func CheckRelationRequest(msg *dsr.CheckRelationRequest) error {
// 	return nil
// }

func GraphRequest(msg *dsr.GraphRequest) error {
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
