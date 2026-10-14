package bdb

type Path []string

// CURRENT definitions.
const (
	manifestName      string = "default"
	modelName         string = "default"
	invertedModelName string = "inverted"
)

var (
	SystemPath         Path = []string{"_system"}
	ManifestPath       Path = ManifestPathV3            // current manifest path
	ManifestPathV3     Path = []string{"manifests"}     // manifests path
	ModelPathV3        Path = []string{"model"}         // model path
	ObjectsPath        Path = []string{"objects"}       // objects path
	RelationsSubPath   Path = []string{"relations_sub"} // relation subject ordered path
	RelationsObjPath   Path = []string{"relations_obj"} // relation object ordered path
	ManifestKeyV3           = []byte(manifestName)
	ModelKeyV3              = []byte(modelName)
	InvertedModelKeyV3      = []byte(invertedModelName)
)

// OBSOLETE definitions.
const (
	manifestVersionV1 string = "0.0.1" // OBSOLETE per migration 0.0.9
)

var (
	ManifestPathV1    Path = []string{"_manifest", manifestName, manifestVersionV1} // OBSOLETE per migration 0.0.8, migration path V1
	ManifestPathV2    Path = []string{"_manifest", manifestName}                    // OBSOLETE per migration 0.0.10, migration path V2
	ObjectTypesPath   Path = []string{"object_types"}                               // OBSOLETE
	PermissionsPath   Path = []string{"permissions"}                                // OBSOLETE
	RelationTypesPath Path = []string{"relation_types"}                             // OBSOLETE
	MetadataKey            = []byte("metadata")                                     // OBSOLETE per migration 0.0.10
	BodyKey                = []byte("body")                                         // OBSOLETE per migration 0.0.10
	ModelKey               = []byte("model")                                        // OBSOLETE per migration 0.0.10
)
