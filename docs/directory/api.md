# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [v3/common/common.proto](#v3_common_common-proto)
    - [Object](#directory-v3-common-Object)
    - [ObjectIdentifier](#directory-v3-common-ObjectIdentifier)
    - [PaginationRequest](#directory-v3-common-PaginationRequest)
    - [PaginationResponse](#directory-v3-common-PaginationResponse)
    - [Relation](#directory-v3-common-Relation)
    - [RelationIdentifier](#directory-v3-common-RelationIdentifier)
  
- [v3/exporter/exporter.proto](#v3_exporter_exporter-proto)
    - [ExportRequest](#directory-v3-exporter-ExportRequest)
    - [ExportResponse](#directory-v3-exporter-ExportResponse)
  
    - [Option](#directory-v3-exporter-Option)
  
    - [Exporter](#directory-v3-exporter-Exporter)
  
- [v3/importer/importer.proto](#v3_importer_importer-proto)
    - [ImportCounter](#directory-v3-importer-ImportCounter)
    - [ImportRequest](#directory-v3-importer-ImportRequest)
    - [ImportResponse](#directory-v3-importer-ImportResponse)
    - [ImportStatus](#directory-v3-importer-ImportStatus)
  
    - [Opcode](#directory-v3-importer-Opcode)
  
    - [Importer](#directory-v3-importer-Importer)
  
- [v3/model/model.proto](#v3_model_model-proto)
    - [Body](#directory-v3-model-Body)
    - [DeleteManifestRequest](#directory-v3-model-DeleteManifestRequest)
    - [DeleteManifestResponse](#directory-v3-model-DeleteManifestResponse)
    - [GetManifestRequest](#directory-v3-model-GetManifestRequest)
    - [GetManifestResponse](#directory-v3-model-GetManifestResponse)
    - [Metadata](#directory-v3-model-Metadata)
    - [SetManifestRequest](#directory-v3-model-SetManifestRequest)
    - [SetManifestResponse](#directory-v3-model-SetManifestResponse)
  
    - [Model](#directory-v3-model-Model)
  
- [v3/openapi/openapi.proto](#v3_openapi_openapi-proto)
- [v3/reader/reader.proto](#v3_reader_reader-proto)
    - [CheckPermissionRequest](#directory-v3-reader-CheckPermissionRequest)
    - [CheckPermissionResponse](#directory-v3-reader-CheckPermissionResponse)
    - [CheckRelationRequest](#directory-v3-reader-CheckRelationRequest)
    - [CheckRelationResponse](#directory-v3-reader-CheckRelationResponse)
    - [CheckRequest](#directory-v3-reader-CheckRequest)
    - [CheckResponse](#directory-v3-reader-CheckResponse)
    - [ChecksRequest](#directory-v3-reader-ChecksRequest)
    - [ChecksResponse](#directory-v3-reader-ChecksResponse)
    - [GetGraphRequest](#directory-v3-reader-GetGraphRequest)
    - [GetGraphResponse](#directory-v3-reader-GetGraphResponse)
    - [GetObjectManyRequest](#directory-v3-reader-GetObjectManyRequest)
    - [GetObjectManyResponse](#directory-v3-reader-GetObjectManyResponse)
    - [GetObjectRequest](#directory-v3-reader-GetObjectRequest)
    - [GetObjectResponse](#directory-v3-reader-GetObjectResponse)
    - [GetObjectsRequest](#directory-v3-reader-GetObjectsRequest)
    - [GetObjectsResponse](#directory-v3-reader-GetObjectsResponse)
    - [GetRelationRequest](#directory-v3-reader-GetRelationRequest)
    - [GetRelationResponse](#directory-v3-reader-GetRelationResponse)
    - [GetRelationResponse.ObjectsEntry](#directory-v3-reader-GetRelationResponse-ObjectsEntry)
    - [GetRelationsRequest](#directory-v3-reader-GetRelationsRequest)
    - [GetRelationsResponse](#directory-v3-reader-GetRelationsResponse)
    - [GetRelationsResponse.ObjectsEntry](#directory-v3-reader-GetRelationsResponse-ObjectsEntry)
  
    - [Reader](#directory-v3-reader-Reader)
  
- [v3/writer/writer.proto](#v3_writer_writer-proto)
    - [DeleteObjectRequest](#directory-v3-writer-DeleteObjectRequest)
    - [DeleteObjectResponse](#directory-v3-writer-DeleteObjectResponse)
    - [DeleteRelationRequest](#directory-v3-writer-DeleteRelationRequest)
    - [DeleteRelationResponse](#directory-v3-writer-DeleteRelationResponse)
    - [SetObjectRequest](#directory-v3-writer-SetObjectRequest)
    - [SetObjectResponse](#directory-v3-writer-SetObjectResponse)
    - [SetRelationRequest](#directory-v3-writer-SetRelationRequest)
    - [SetRelationResponse](#directory-v3-writer-SetRelationResponse)
  
    - [Writer](#directory-v3-writer-Writer)
  
- [Scalar Value Types](#scalar-value-types)



<a name="v3_common_common-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## v3/common/common.proto



<a name="directory-v3-common-Object"></a>

### Object



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [string](#string) |  | object type identifier |
| id | [string](#string) |  | object instance identifier |
| display_name | [string](#string) |  | **Deprecated.** display name object (optional) (DEPRECATED) |
| properties | [google.protobuf.Struct](#google-protobuf-Struct) |  | property bag (optional) |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | **Deprecated.** created at timestamp (UTC) (DEPRECATED) |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | last updated timestamp (UTC) |
| etag | [string](#string) |  | object instance etag (optional) |






<a name="directory-v3-common-ObjectIdentifier"></a>

### ObjectIdentifier
Object identifier


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |






<a name="directory-v3-common-PaginationRequest"></a>

### PaginationRequest
Pagination request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size | [int32](#int32) |  | requested page size, valid value between 1-100 rows (optional, default 100) |
| token | [string](#string) |  | pagination start token (optional, default &#34;&#34;) |






<a name="directory-v3-common-PaginationResponse"></a>

### PaginationResponse
Pagination response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| next_token | [string](#string) |  | next page token, when empty there are no more pages to fetch |






<a name="directory-v3-common-Relation"></a>

### Relation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |
| relation | [string](#string) |  | object relation name |
| subject_type | [string](#string) |  | subject type identifier |
| subject_id | [string](#string) |  | subject instance identifier |
| subject_relation | [string](#string) |  | subject relation name (optional) |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | **Deprecated.** created at timestamp (UTC) (DEPRECATED) |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | last updated timestamp (UTC) |
| etag | [string](#string) |  | object instance etag (optional) |






<a name="directory-v3-common-RelationIdentifier"></a>

### RelationIdentifier
Relation identifier


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |
| relation | [string](#string) |  | object relation name |
| subject_type | [string](#string) |  | subject type identifier |
| subject_id | [string](#string) |  | subject instance identifier |
| subject_relation | [string](#string) |  | subject relation name (optional) |





 

 

 

 



<a name="v3_exporter_exporter-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## v3/exporter/exporter.proto



<a name="directory-v3-exporter-ExportRequest"></a>

### ExportRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| options | [uint32](#uint32) |  | data export options mask |
| start_from | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | start export from timestamp (UTC) |






<a name="directory-v3-exporter-ExportResponse"></a>

### ExportResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object | [directory.v3.common.Object](#directory-v3-common-Object) |  | object instance (data) |
| relation | [directory.v3.common.Relation](#directory-v3-common-Relation) |  | relation instance (data) |
| stats | [google.protobuf.Struct](#google-protobuf-Struct) |  | object and/or relation stats (no data) |





 


<a name="directory-v3-exporter-Option"></a>

### Option


| Name | Number | Description |
| ---- | ------ | ----------- |
| OPTION_UNKNOWN | 0 | nothing selected (default initialization value) |
| OPTION_DATA_OBJECTS | 8 | object instances |
| OPTION_DATA_RELATIONS | 16 | relation instances |
| OPTION_DATA | 24 | all data = OPTION_DATA_OBJECTS | OPTION_DATA_RELATIONS |
| OPTION_STATS | 64 | stats |


 

 


<a name="directory-v3-exporter-Exporter"></a>

### Exporter


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Export | [ExportRequest](#directory-v3-exporter-ExportRequest) | [ExportResponse](#directory-v3-exporter-ExportResponse) stream | export objects and relations as a stream |

 



<a name="v3_importer_importer-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## v3/importer/importer.proto



<a name="directory-v3-importer-ImportCounter"></a>

### ImportCounter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| recv | [uint64](#uint64) |  | number of messages received |
| set | [uint64](#uint64) |  | number of messages with OPCODE_SET |
| delete | [uint64](#uint64) |  | number of messages with OPCODE_DELETE |
| error | [uint64](#uint64) |  | number of messages resulting in error |
| type | [string](#string) |  | counter of type (object|relation) |






<a name="directory-v3-importer-ImportRequest"></a>

### ImportRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| op_code | [Opcode](#directory-v3-importer-Opcode) |  | operation Opcode enum value |
| object | [directory.v3.common.Object](#directory-v3-common-Object) |  | object import message |
| relation | [directory.v3.common.Relation](#directory-v3-common-Relation) |  | relation import message |






<a name="directory-v3-importer-ImportResponse"></a>

### ImportResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object | [ImportCounter](#directory-v3-importer-ImportCounter) |  | **Deprecated.** object import counter |
| relation | [ImportCounter](#directory-v3-importer-ImportCounter) |  | **Deprecated.** relation import counter |
| status | [ImportStatus](#directory-v3-importer-ImportStatus) |  | import status message |
| counter | [ImportCounter](#directory-v3-importer-ImportCounter) |  | import counter per type |






<a name="directory-v3-importer-ImportStatus"></a>

### ImportStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [uint32](#uint32) |  | gRPC status code (google.golang.org/grpc/codes) |
| msg | [string](#string) |  | gRPC status message (google.golang.org/grpc/status) |
| req | [ImportRequest](#directory-v3-importer-ImportRequest) |  | req contains the original import request message |





 


<a name="directory-v3-importer-Opcode"></a>

### Opcode


| Name | Number | Description |
| ---- | ------ | ----------- |
| OPCODE_UNKNOWN | 0 |  |
| OPCODE_SET | 1 |  |
| OPCODE_DELETE | 2 |  |
| OPCODE_DELETE_WITH_RELATIONS | 3 |  |


 

 


<a name="directory-v3-importer-Importer"></a>

### Importer


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Import | [ImportRequest](#directory-v3-importer-ImportRequest) stream | [ImportResponse](#directory-v3-importer-ImportResponse) stream | import stream of objects and relations |

 



<a name="v3_model_model-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## v3/model/model.proto



<a name="directory-v3-model-Body"></a>

### Body



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| data | [bytes](#bytes) |  | manifest content |






<a name="directory-v3-model-DeleteManifestRequest"></a>

### DeleteManifestRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| empty | [google.protobuf.Empty](#google-protobuf-Empty) |  | empty request |






<a name="directory-v3-model-DeleteManifestResponse"></a>

### DeleteManifestResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [google.protobuf.Empty](#google-protobuf-Empty) |  | empty result |






<a name="directory-v3-model-GetManifestRequest"></a>

### GetManifestRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| empty | [google.protobuf.Empty](#google-protobuf-Empty) |  | empty request |






<a name="directory-v3-model-GetManifestResponse"></a>

### GetManifestResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [Metadata](#directory-v3-model-Metadata) |  | Manifest metadata |
| body | [Body](#directory-v3-model-Body) |  | Manifest content |
| model | [google.protobuf.Struct](#google-protobuf-Struct) |  | Model |






<a name="directory-v3-model-Metadata"></a>

### Metadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | last updated timestamp (UTC) |
| etag | [string](#string) |  | object instance etag (optional) |






<a name="directory-v3-model-SetManifestRequest"></a>

### SetManifestRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| body | [Body](#directory-v3-model-Body) |  | Manifest content |






<a name="directory-v3-model-SetManifestResponse"></a>

### SetManifestResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [google.protobuf.Empty](#google-protobuf-Empty) |  | empty result |





 

 

 


<a name="directory-v3-model-Model"></a>

### Model


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetManifest | [GetManifestRequest](#directory-v3-model-GetManifestRequest) | [GetManifestResponse](#directory-v3-model-GetManifestResponse) stream | get manifest instance |
| SetManifest | [SetManifestRequest](#directory-v3-model-SetManifestRequest) stream | [SetManifestResponse](#directory-v3-model-SetManifestResponse) | set manifest instance |
| DeleteManifest | [DeleteManifestRequest](#directory-v3-model-DeleteManifestRequest) | [DeleteManifestResponse](#directory-v3-model-DeleteManifestResponse) | delete manifest instance |

 



<a name="v3_openapi_openapi-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## v3/openapi/openapi.proto


 

 

 

 



<a name="v3_reader_reader-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## v3/reader/reader.proto



<a name="directory-v3-reader-CheckPermissionRequest"></a>

### CheckPermissionRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |
| permission | [string](#string) |  | permission name |
| subject_type | [string](#string) |  | subject type identifier |
| subject_id | [string](#string) |  | subject instance identifier |
| trace | [bool](#bool) |  | collect trace information (optional) |






<a name="directory-v3-reader-CheckPermissionResponse"></a>

### CheckPermissionResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| check | [bool](#bool) |  | check result |
| trace | [string](#string) | repeated | trace information |






<a name="directory-v3-reader-CheckRelationRequest"></a>

### CheckRelationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |
| relation | [string](#string) |  | relation name |
| subject_type | [string](#string) |  | subject type identifier |
| subject_id | [string](#string) |  | subject instance identifier |
| trace | [bool](#bool) |  | collect trace information (optional) |






<a name="directory-v3-reader-CheckRelationResponse"></a>

### CheckRelationResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| check | [bool](#bool) |  | check result |
| trace | [string](#string) | repeated | trace information |






<a name="directory-v3-reader-CheckRequest"></a>

### CheckRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |
| relation | [string](#string) |  | relation name |
| subject_type | [string](#string) |  | subject type identifier |
| subject_id | [string](#string) |  | subject instance identifier |
| trace | [bool](#bool) |  | collect trace information (optional) |






<a name="directory-v3-reader-CheckResponse"></a>

### CheckResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| check | [bool](#bool) |  | check result |
| trace | [string](#string) | repeated | trace information |
| context | [google.protobuf.Struct](#google-protobuf-Struct) |  | context |






<a name="directory-v3-reader-ChecksRequest"></a>

### ChecksRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| default | [CheckRequest](#directory-v3-reader-CheckRequest) |  | default check parameter values |
| checks | [CheckRequest](#directory-v3-reader-CheckRequest) | repeated | collection of check requests |






<a name="directory-v3-reader-ChecksResponse"></a>

### ChecksResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| checks | [CheckResponse](#directory-v3-reader-CheckResponse) | repeated | checks results array |






<a name="directory-v3-reader-GetGraphRequest"></a>

### GetGraphRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier (optional) |
| relation | [string](#string) |  | relation name |
| subject_type | [string](#string) |  | subject type identifier |
| subject_id | [string](#string) |  | subject instance identifier (optional) |
| subject_relation | [string](#string) |  | subject relation name (optional) |
| explain | [bool](#bool) |  | return graph paths for each result (optional) |
| trace | [bool](#bool) |  | collect trace information (optional) |






<a name="directory-v3-reader-GetGraphResponse"></a>

### GetGraphResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| results | [directory.v3.common.ObjectIdentifier](#directory-v3-common-ObjectIdentifier) | repeated | matching object identifiers |
| explanation | [google.protobuf.Struct](#google-protobuf-Struct) |  | explanation of results |
| trace | [string](#string) | repeated | trace information |






<a name="directory-v3-reader-GetObjectManyRequest"></a>

### GetObjectManyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| param | [directory.v3.common.ObjectIdentifier](#directory-v3-common-ObjectIdentifier) | repeated | object identifier list |






<a name="directory-v3-reader-GetObjectManyResponse"></a>

### GetObjectManyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| results | [directory.v3.common.Object](#directory-v3-common-Object) | repeated | array of object instances |






<a name="directory-v3-reader-GetObjectRequest"></a>

### GetObjectRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |
| with_relations | [bool](#bool) |  | materialize the object relations objects (optional) |






<a name="directory-v3-reader-GetObjectResponse"></a>

### GetObjectResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [directory.v3.common.Object](#directory-v3-common-Object) |  | object instance |
| relations | [directory.v3.common.Relation](#directory-v3-common-Relation) | repeated | object relations |






<a name="directory-v3-reader-GetObjectsRequest"></a>

### GetObjectsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier (optional) |
| page | [directory.v3.common.PaginationRequest](#directory-v3-common-PaginationRequest) |  | pagination request (optional) |






<a name="directory-v3-reader-GetObjectsResponse"></a>

### GetObjectsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| results | [directory.v3.common.Object](#directory-v3-common-Object) | repeated | array of object instances |
| page | [directory.v3.common.PaginationResponse](#directory-v3-common-PaginationResponse) |  | pagination response |






<a name="directory-v3-reader-GetRelationRequest"></a>

### GetRelationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |
| relation | [string](#string) |  | relation name |
| subject_type | [string](#string) |  | subject type identifier |
| subject_id | [string](#string) |  | subject instance identifier |
| subject_relation | [string](#string) |  | subject relation name (optional) |
| with_objects | [bool](#bool) |  | materialize relation objects (optional) |






<a name="directory-v3-reader-GetRelationResponse"></a>

### GetRelationResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [directory.v3.common.Relation](#directory-v3-common-Relation) |  | relation instance |
| objects | [GetRelationResponse.ObjectsEntry](#directory-v3-reader-GetRelationResponse-ObjectsEntry) | repeated | map of materialized relation objects |






<a name="directory-v3-reader-GetRelationResponse-ObjectsEntry"></a>

### GetRelationResponse.ObjectsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [directory.v3.common.Object](#directory-v3-common-Object) |  |  |






<a name="directory-v3-reader-GetRelationsRequest"></a>

### GetRelationsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier (optional) |
| object_id | [string](#string) |  | object instance identifier (optional) |
| relation | [string](#string) |  | relation name (optional) |
| subject_type | [string](#string) |  | subject type identifier (optional) |
| subject_id | [string](#string) |  | subject instance identifier (optional) |
| subject_relation | [string](#string) |  | subject relation name (optional) |
| with_objects | [bool](#bool) |  | materialize relation objects (optional) |
| with_empty_subject_relation | [bool](#bool) |  | only return relations that do not have a subject relation (optional) |
| page | [directory.v3.common.PaginationRequest](#directory-v3-common-PaginationRequest) |  | pagination request (optional) |






<a name="directory-v3-reader-GetRelationsResponse"></a>

### GetRelationsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| results | [directory.v3.common.Relation](#directory-v3-common-Relation) | repeated | array of relation instances |
| objects | [GetRelationsResponse.ObjectsEntry](#directory-v3-reader-GetRelationsResponse-ObjectsEntry) | repeated | map of materialized relation objects |
| page | [directory.v3.common.PaginationResponse](#directory-v3-common-PaginationResponse) |  | pagination response |






<a name="directory-v3-reader-GetRelationsResponse-ObjectsEntry"></a>

### GetRelationsResponse.ObjectsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [directory.v3.common.Object](#directory-v3-common-Object) |  |  |





 

 

 


<a name="directory-v3-reader-Reader"></a>

### Reader


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetObject | [GetObjectRequest](#directory-v3-reader-GetObjectRequest) | [GetObjectResponse](#directory-v3-reader-GetObjectResponse) | get object |
| GetObjectMany | [GetObjectManyRequest](#directory-v3-reader-GetObjectManyRequest) | [GetObjectManyResponse](#directory-v3-reader-GetObjectManyResponse) | get multiple objects |
| GetObjects | [GetObjectsRequest](#directory-v3-reader-GetObjectsRequest) | [GetObjectsResponse](#directory-v3-reader-GetObjectsResponse) | list objects |
| GetRelation | [GetRelationRequest](#directory-v3-reader-GetRelationRequest) | [GetRelationResponse](#directory-v3-reader-GetRelationResponse) | get relation |
| GetRelations | [GetRelationsRequest](#directory-v3-reader-GetRelationsRequest) | [GetRelationsResponse](#directory-v3-reader-GetRelationsResponse) | list relations |
| Check | [CheckRequest](#directory-v3-reader-CheckRequest) | [CheckResponse](#directory-v3-reader-CheckResponse) | check if subject has relation or permission with object |
| Checks | [ChecksRequest](#directory-v3-reader-ChecksRequest) | [ChecksResponse](#directory-v3-reader-ChecksResponse) | checks validates a set of check requests in a single roundtrip |
| CheckPermission | [CheckPermissionRequest](#directory-v3-reader-CheckPermissionRequest) | [CheckPermissionResponse](#directory-v3-reader-CheckPermissionResponse) | check permission (deprecated, use the check method) Deprecated: use directory.reader.v3.Check() |
| CheckRelation | [CheckRelationRequest](#directory-v3-reader-CheckRelationRequest) | [CheckRelationResponse](#directory-v3-reader-CheckRelationResponse) | check relation (deprecated, use the check method) Deprecated: use directory.reader.v3.Check() |
| GetGraph | [GetGraphRequest](#directory-v3-reader-GetGraphRequest) | [GetGraphResponse](#directory-v3-reader-GetGraphResponse) | get object relationship graph |

 



<a name="v3_writer_writer-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## v3/writer/writer.proto



<a name="directory-v3-writer-DeleteObjectRequest"></a>

### DeleteObjectRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |
| with_relations | [bool](#bool) |  | delete object relations, both object and subject relations (optional) |






<a name="directory-v3-writer-DeleteObjectResponse"></a>

### DeleteObjectResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [google.protobuf.Empty](#google-protobuf-Empty) |  | empty result |






<a name="directory-v3-writer-DeleteRelationRequest"></a>

### DeleteRelationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |
| relation | [string](#string) |  | object relation name |
| subject_type | [string](#string) |  | subject type identifier |
| subject_id | [string](#string) |  | subject instance identifier |
| subject_relation | [string](#string) |  | subject relation name (optional) |






<a name="directory-v3-writer-DeleteRelationResponse"></a>

### DeleteRelationResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [google.protobuf.Empty](#google-protobuf-Empty) |  | empty result |






<a name="directory-v3-writer-SetObjectRequest"></a>

### SetObjectRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object | [directory.v3.common.Object](#directory-v3-common-Object) |  | object instance |






<a name="directory-v3-writer-SetObjectResponse"></a>

### SetObjectResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [directory.v3.common.Object](#directory-v3-common-Object) |  | object instance |






<a name="directory-v3-writer-SetRelationRequest"></a>

### SetRelationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| relation | [directory.v3.common.Relation](#directory-v3-common-Relation) |  | relation instance |






<a name="directory-v3-writer-SetRelationResponse"></a>

### SetRelationResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [directory.v3.common.Relation](#directory-v3-common-Relation) |  | relation instance |





 

 

 


<a name="directory-v3-writer-Writer"></a>

### Writer


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| SetObject | [SetObjectRequest](#directory-v3-writer-SetObjectRequest) | [SetObjectResponse](#directory-v3-writer-SetObjectResponse) | set object instance |
| DeleteObject | [DeleteObjectRequest](#directory-v3-writer-DeleteObjectRequest) | [DeleteObjectResponse](#directory-v3-writer-DeleteObjectResponse) | delete object instance |
| SetRelation | [SetRelationRequest](#directory-v3-writer-SetRelationRequest) | [SetRelationResponse](#directory-v3-writer-SetRelationResponse) | set relation instance |
| DeleteRelation | [DeleteRelationRequest](#directory-v3-writer-DeleteRelationRequest) | [DeleteRelationResponse](#directory-v3-writer-DeleteRelationResponse) | delete relation instance |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

