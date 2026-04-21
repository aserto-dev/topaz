package validator

import (
	dsc "github.com/aserto-dev/topaz/api/directory/v4"
	dsr "github.com/aserto-dev/topaz/api/directory/v4/reader"
	dsw "github.com/aserto-dev/topaz/api/directory/v4/writer"
)

func Object(msg *dsc.Object) error {
	if msg == nil {
		return nil
	}

	if err := TypeIdentifier(fieldType, msg.GetObjectType()); err != nil {
		return err
	}

	if err := InstanceIdentifier(fieldID, msg.GetObjectId()); err != nil {
		return err
	}

	if err := Etag(fieldETag, msg.GetEtag()); err != nil {
		return err
	}

	return nil
}

func ObjectIdentifier(msg *dsc.ObjectIdentifier) error {
	if msg == nil {
		return nil
	}

	if err := TypeIdentifier(fieldObjectType, msg.GetObjectType()); err != nil {
		return err
	}

	if err := InstanceIdentifier(fieldObjectID, msg.GetObjectId()); err != nil {
		return err
	}

	return nil
}

func GetObjectRequest(msg *dsr.GetObjectRequest) error {
	if msg == nil {
		return nil
	}

	if err := TypeIdentifier(fieldObjectType, msg.GetObjectType()); err != nil {
		return err
	}

	if err := InstanceIdentifier(fieldObjectID, msg.GetObjectId()); err != nil {
		return err
	}

	return nil
}

func ListObjectsRequest(msg *dsr.ListObjectsRequest) error {
	if msg == nil {
		return nil
	}

	if err := TypeIdentifier(fieldObjectType, msg.GetObjectType()); err != nil {
		return err
	}

	if err := PaginationRequest(msg.GetPage()); err != nil {
		return err
	}

	return nil
}

// func GetObjectManyRequest(msg *dsr.GetObjectManyRequest) error {
// 	if msg == nil || msg.GetParam() == nil || len(msg.GetParam()) == 0 {
// 		return nil
// 	}

// 	for _, oid := range msg.GetParam() {
// 		if err := ObjectIdentifier(oid); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

func SetObjectRequest(msg *dsw.SetObjectRequest) error {
	if msg == nil || msg.GetObject() == nil {
		return nil
	}

	if err := Object(msg.GetObject()); err != nil {
		return err
	}

	return nil
}

func DeleteObjectRequest(msg *dsw.DeleteObjectRequest) error {
	if msg == nil {
		return nil
	}

	if err := TypeIdentifier(fieldObjectType, msg.GetObjectType()); err != nil {
		return err
	}

	if err := InstanceIdentifier(fieldObjectID, msg.GetObjectId()); err != nil {
		return err
	}

	return nil
}
