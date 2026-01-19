package bdb

type Path []string

const (
	manifestName      string = "default"
	manifestVersionV1 string = "0.0.1" // OBSOLETE per migration 0.0.9
)

var (
	SystemPath        Path = []string{"_system"}
	ManifestPath      Path = ManifestPathV2                                         // current path
	ManifestPathV1    Path = []string{"_manifest", manifestName, manifestVersionV1} // migration path V1, OBSOLETE per migration 0.0.8
	ManifestPathV2    Path = []string{"_manifest", manifestName}                    // migration path V2
	ObjectTypesPath   Path = []string{"object_types"}                               // OBSOLETE
	PermissionsPath   Path = []string{"permissions"}                                // OBSOLETE
	RelationTypesPath Path = []string{"relation_types"}                             // OBSOLETE
	ObjectsPath       Path = []string{"objects"}
	RelationsSubPath  Path = []string{"relations_sub"}
	RelationsObjPath  Path = []string{"relations_obj"}
	MetadataKey            = []byte("metadata")
	BodyKey                = []byte("body")
	ModelKey               = []byte("model")
)
