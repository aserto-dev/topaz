# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [v3/common/common.proto](#v3_common_common-proto)
    - [Object](#aserto-directory-common-v3-Object)
    - [ObjectIdentifier](#aserto-directory-common-v3-ObjectIdentifier)
    - [PaginationRequest](#aserto-directory-common-v3-PaginationRequest)
    - [PaginationResponse](#aserto-directory-common-v3-PaginationResponse)
    - [Relation](#aserto-directory-common-v3-Relation)
    - [RelationIdentifier](#aserto-directory-common-v3-RelationIdentifier)
  
- [v3/exporter/exporter.proto](#v3_exporter_exporter-proto)
    - [ExportRequest](#aserto-directory-exporter-v3-ExportRequest)
    - [ExportResponse](#aserto-directory-exporter-v3-ExportResponse)
  
    - [Option](#aserto-directory-exporter-v3-Option)
  
    - [Exporter](#aserto-directory-exporter-v3-Exporter)
  
- [v3/importer/importer.proto](#v3_importer_importer-proto)
    - [ImportCounter](#aserto-directory-importer-v3-ImportCounter)
    - [ImportRequest](#aserto-directory-importer-v3-ImportRequest)
    - [ImportResponse](#aserto-directory-importer-v3-ImportResponse)
    - [ImportStatus](#aserto-directory-importer-v3-ImportStatus)
  
    - [Opcode](#aserto-directory-importer-v3-Opcode)
  
    - [Importer](#aserto-directory-importer-v3-Importer)
  
- [v3/model/model.proto](#v3_model_model-proto)
    - [Body](#aserto-directory-model-v3-Body)
    - [DeleteManifestRequest](#aserto-directory-model-v3-DeleteManifestRequest)
    - [DeleteManifestResponse](#aserto-directory-model-v3-DeleteManifestResponse)
    - [GetManifestRequest](#aserto-directory-model-v3-GetManifestRequest)
    - [GetManifestResponse](#aserto-directory-model-v3-GetManifestResponse)
    - [Metadata](#aserto-directory-model-v3-Metadata)
    - [SetManifestRequest](#aserto-directory-model-v3-SetManifestRequest)
    - [SetManifestResponse](#aserto-directory-model-v3-SetManifestResponse)
  
    - [Model](#aserto-directory-model-v3-Model)
  
- [v3/openapi/openapi.proto](#v3_openapi_openapi-proto)
- [v3/reader/reader.proto](#v3_reader_reader-proto)
    - [CheckPermissionRequest](#aserto-directory-reader-v3-CheckPermissionRequest)
    - [CheckPermissionResponse](#aserto-directory-reader-v3-CheckPermissionResponse)
    - [CheckRelationRequest](#aserto-directory-reader-v3-CheckRelationRequest)
    - [CheckRelationResponse](#aserto-directory-reader-v3-CheckRelationResponse)
    - [CheckRequest](#aserto-directory-reader-v3-CheckRequest)
    - [CheckResponse](#aserto-directory-reader-v3-CheckResponse)
    - [ChecksRequest](#aserto-directory-reader-v3-ChecksRequest)
    - [ChecksResponse](#aserto-directory-reader-v3-ChecksResponse)
    - [GetGraphRequest](#aserto-directory-reader-v3-GetGraphRequest)
    - [GetGraphResponse](#aserto-directory-reader-v3-GetGraphResponse)
    - [GetObjectManyRequest](#aserto-directory-reader-v3-GetObjectManyRequest)
    - [GetObjectManyResponse](#aserto-directory-reader-v3-GetObjectManyResponse)
    - [GetObjectRequest](#aserto-directory-reader-v3-GetObjectRequest)
    - [GetObjectResponse](#aserto-directory-reader-v3-GetObjectResponse)
    - [GetObjectsRequest](#aserto-directory-reader-v3-GetObjectsRequest)
    - [GetObjectsResponse](#aserto-directory-reader-v3-GetObjectsResponse)
    - [GetRelationRequest](#aserto-directory-reader-v3-GetRelationRequest)
    - [GetRelationResponse](#aserto-directory-reader-v3-GetRelationResponse)
    - [GetRelationResponse.ObjectsEntry](#aserto-directory-reader-v3-GetRelationResponse-ObjectsEntry)
    - [GetRelationsRequest](#aserto-directory-reader-v3-GetRelationsRequest)
    - [GetRelationsResponse](#aserto-directory-reader-v3-GetRelationsResponse)
    - [GetRelationsResponse.ObjectsEntry](#aserto-directory-reader-v3-GetRelationsResponse-ObjectsEntry)
  
    - [Reader](#aserto-directory-reader-v3-Reader)
  
- [v3/writer/writer.proto](#v3_writer_writer-proto)
    - [DeleteObjectRequest](#aserto-directory-writer-v3-DeleteObjectRequest)
    - [DeleteObjectResponse](#aserto-directory-writer-v3-DeleteObjectResponse)
    - [DeleteRelationRequest](#aserto-directory-writer-v3-DeleteRelationRequest)
    - [DeleteRelationResponse](#aserto-directory-writer-v3-DeleteRelationResponse)
    - [SetObjectRequest](#aserto-directory-writer-v3-SetObjectRequest)
    - [SetObjectResponse](#aserto-directory-writer-v3-SetObjectResponse)
    - [SetRelationRequest](#aserto-directory-writer-v3-SetRelationRequest)
    - [SetRelationResponse](#aserto-directory-writer-v3-SetRelationResponse)
  
    - [Writer](#aserto-directory-writer-v3-Writer)
  
- [Scalar Value Types](#scalar-value-types)



<a name="v3_common_common-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## v3/common/common.proto



<a name="aserto-directory-common-v3-Object"></a>

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






<a name="aserto-directory-common-v3-ObjectIdentifier"></a>

### ObjectIdentifier
Object identifier


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |






<a name="aserto-directory-common-v3-PaginationRequest"></a>

### PaginationRequest
Pagination request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size | [int32](#int32) |  | requested page size, valid value between 1-100 rows (optional, default 100) |
| token | [string](#string) |  | pagination start token (optional, default &#34;&#34;) |






<a name="aserto-directory-common-v3-PaginationResponse"></a>

### PaginationResponse
Pagination response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| next_token | [string](#string) |  | next page token, when empty there are no more pages to fetch |






<a name="aserto-directory-common-v3-Relation"></a>

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






<a name="aserto-directory-common-v3-RelationIdentifier"></a>

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



<a name="aserto-directory-exporter-v3-ExportRequest"></a>

### ExportRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| options | [uint32](#uint32) |  | data export options mask |
| start_from | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | start export from timestamp (UTC) |






<a name="aserto-directory-exporter-v3-ExportResponse"></a>

### ExportResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object | [aserto.directory.common.v3.Object](#aserto-directory-common-v3-Object) |  | object instance (data) |
| relation | [aserto.directory.common.v3.Relation](#aserto-directory-common-v3-Relation) |  | relation instance (data) |
| stats | [google.protobuf.Struct](#google-protobuf-Struct) |  | object and/or relation stats (no data) |





 


<a name="aserto-directory-exporter-v3-Option"></a>

### Option


| Name | Number | Description |
| ---- | ------ | ----------- |
| OPTION_UNKNOWN | 0 | nothing selected (default initialization value) |
| OPTION_DATA_OBJECTS | 8 | object instances |
| OPTION_DATA_RELATIONS | 16 | relation instances |
| OPTION_DATA | 24 | all data = OPTION_DATA_OBJECTS | OPTION_DATA_RELATIONS |
| OPTION_STATS | 64 | stats |


 

 


<a name="aserto-directory-exporter-v3-Exporter"></a>

### Exporter


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Export | [ExportRequest](#aserto-directory-exporter-v3-ExportRequest) | [ExportResponse](#aserto-directory-exporter-v3-ExportResponse) stream | export objects and relations as a stream |

 



<a name="v3_importer_importer-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## v3/importer/importer.proto



<a name="aserto-directory-importer-v3-ImportCounter"></a>

### ImportCounter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| recv | [uint64](#uint64) |  | number of messages received |
| set | [uint64](#uint64) |  | number of messages with OPCODE_SET |
| delete | [uint64](#uint64) |  | number of messages with OPCODE_DELETE |
| error | [uint64](#uint64) |  | number of messages resulting in error |
| type | [string](#string) |  | counter of type (object|relation) |






<a name="aserto-directory-importer-v3-ImportRequest"></a>

### ImportRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| op_code | [Opcode](#aserto-directory-importer-v3-Opcode) |  | operation Opcode enum value |
| object | [aserto.directory.common.v3.Object](#aserto-directory-common-v3-Object) |  | object import message |
| relation | [aserto.directory.common.v3.Relation](#aserto-directory-common-v3-Relation) |  | relation import message |






<a name="aserto-directory-importer-v3-ImportResponse"></a>

### ImportResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object | [ImportCounter](#aserto-directory-importer-v3-ImportCounter) |  | **Deprecated.** object import counter |
| relation | [ImportCounter](#aserto-directory-importer-v3-ImportCounter) |  | **Deprecated.** relation import counter |
| status | [ImportStatus](#aserto-directory-importer-v3-ImportStatus) |  | import status message |
| counter | [ImportCounter](#aserto-directory-importer-v3-ImportCounter) |  | import counter per type |






<a name="aserto-directory-importer-v3-ImportStatus"></a>

### ImportStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [uint32](#uint32) |  | gRPC status code (google.golang.org/grpc/codes) |
| msg | [string](#string) |  | gRPC status message (google.golang.org/grpc/status) |
| req | [ImportRequest](#aserto-directory-importer-v3-ImportRequest) |  | req contains the original import request message |





 


<a name="aserto-directory-importer-v3-Opcode"></a>

### Opcode


| Name | Number | Description |
| ---- | ------ | ----------- |
| OPCODE_UNKNOWN | 0 |  |
| OPCODE_SET | 1 |  |
| OPCODE_DELETE | 2 |  |
| OPCODE_DELETE_WITH_RELATIONS | 3 |  |


 

 


<a name="aserto-directory-importer-v3-Importer"></a>

### Importer


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Import | [ImportRequest](#aserto-directory-importer-v3-ImportRequest) stream | [ImportResponse](#aserto-directory-importer-v3-ImportResponse) stream | import stream of objects and relations |

 



<a name="v3_model_model-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## v3/model/model.proto



<a name="aserto-directory-model-v3-Body"></a>

### Body



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| data | [bytes](#bytes) |  | manifest content |






<a name="aserto-directory-model-v3-DeleteManifestRequest"></a>

### DeleteManifestRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| empty | [google.protobuf.Empty](#google-protobuf-Empty) |  | empty request |






<a name="aserto-directory-model-v3-DeleteManifestResponse"></a>

### DeleteManifestResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [google.protobuf.Empty](#google-protobuf-Empty) |  | empty result |






<a name="aserto-directory-model-v3-GetManifestRequest"></a>

### GetManifestRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| empty | [google.protobuf.Empty](#google-protobuf-Empty) |  | empty request |






<a name="aserto-directory-model-v3-GetManifestResponse"></a>

### GetManifestResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [Metadata](#aserto-directory-model-v3-Metadata) |  | Manifest metadata |
| body | [Body](#aserto-directory-model-v3-Body) |  | Manifest content |
| model | [google.protobuf.Struct](#google-protobuf-Struct) |  | Model |






<a name="aserto-directory-model-v3-Metadata"></a>

### Metadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | last updated timestamp (UTC) |
| etag | [string](#string) |  | object instance etag (optional) |






<a name="aserto-directory-model-v3-SetManifestRequest"></a>

### SetManifestRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| body | [Body](#aserto-directory-model-v3-Body) |  | Manifest content |






<a name="aserto-directory-model-v3-SetManifestResponse"></a>

### SetManifestResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [google.protobuf.Empty](#google-protobuf-Empty) |  | empty result |





 

 

 


<a name="aserto-directory-model-v3-Model"></a>

### Model


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetManifest | [GetManifestRequest](#aserto-directory-model-v3-GetManifestRequest) | [GetManifestResponse](#aserto-directory-model-v3-GetManifestResponse) stream | get manifest instance |
| SetManifest | [SetManifestRequest](#aserto-directory-model-v3-SetManifestRequest) stream | [SetManifestResponse](#aserto-directory-model-v3-SetManifestResponse) | set manifest instance |
| DeleteManifest | [DeleteManifestRequest](#aserto-directory-model-v3-DeleteManifestRequest) | [DeleteManifestResponse](#aserto-directory-model-v3-DeleteManifestResponse) | delete manifest instance |

 



<a name="v3_openapi_openapi-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## v3/openapi/openapi.proto


 

 

 

 



<a name="v3_reader_reader-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## v3/reader/reader.proto



<a name="aserto-directory-reader-v3-CheckPermissionRequest"></a>

### CheckPermissionRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |
| permission | [string](#string) |  | permission name |
| subject_type | [string](#string) |  | subject type identifier |
| subject_id | [string](#string) |  | subject instance identifier |
| trace | [bool](#bool) |  | collect trace information (optional) |






<a name="aserto-directory-reader-v3-CheckPermissionResponse"></a>

### CheckPermissionResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| check | [bool](#bool) |  | check result |
| trace | [string](#string) | repeated | trace information |






<a name="aserto-directory-reader-v3-CheckRelationRequest"></a>

### CheckRelationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |
| relation | [string](#string) |  | relation name |
| subject_type | [string](#string) |  | subject type identifier |
| subject_id | [string](#string) |  | subject instance identifier |
| trace | [bool](#bool) |  | collect trace information (optional) |






<a name="aserto-directory-reader-v3-CheckRelationResponse"></a>

### CheckRelationResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| check | [bool](#bool) |  | check result |
| trace | [string](#string) | repeated | trace information |






<a name="aserto-directory-reader-v3-CheckRequest"></a>

### CheckRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |
| relation | [string](#string) |  | relation name |
| subject_type | [string](#string) |  | subject type identifier |
| subject_id | [string](#string) |  | subject instance identifier |
| trace | [bool](#bool) |  | collect trace information (optional) |






<a name="aserto-directory-reader-v3-CheckResponse"></a>

### CheckResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| check | [bool](#bool) |  | check result |
| trace | [string](#string) | repeated | trace information |
| context | [google.protobuf.Struct](#google-protobuf-Struct) |  | context |






<a name="aserto-directory-reader-v3-ChecksRequest"></a>

### ChecksRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| default | [CheckRequest](#aserto-directory-reader-v3-CheckRequest) |  | default check parameter values |
| checks | [CheckRequest](#aserto-directory-reader-v3-CheckRequest) | repeated | collection of check requests |






<a name="aserto-directory-reader-v3-ChecksResponse"></a>

### ChecksResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| checks | [CheckResponse](#aserto-directory-reader-v3-CheckResponse) | repeated | checks results array |






<a name="aserto-directory-reader-v3-GetGraphRequest"></a>

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






<a name="aserto-directory-reader-v3-GetGraphResponse"></a>

### GetGraphResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| results | [aserto.directory.common.v3.ObjectIdentifier](#aserto-directory-common-v3-ObjectIdentifier) | repeated | matching object identifiers |
| explanation | [google.protobuf.Struct](#google-protobuf-Struct) |  | explanation of results |
| trace | [string](#string) | repeated | trace information |






<a name="aserto-directory-reader-v3-GetObjectManyRequest"></a>

### GetObjectManyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| param | [aserto.directory.common.v3.ObjectIdentifier](#aserto-directory-common-v3-ObjectIdentifier) | repeated | object identifier list |






<a name="aserto-directory-reader-v3-GetObjectManyResponse"></a>

### GetObjectManyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| results | [aserto.directory.common.v3.Object](#aserto-directory-common-v3-Object) | repeated | array of object instances |






<a name="aserto-directory-reader-v3-GetObjectRequest"></a>

### GetObjectRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |
| with_relations | [bool](#bool) |  | materialize the object relations objects (optional) |






<a name="aserto-directory-reader-v3-GetObjectResponse"></a>

### GetObjectResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [aserto.directory.common.v3.Object](#aserto-directory-common-v3-Object) |  | object instance |
| relations | [aserto.directory.common.v3.Relation](#aserto-directory-common-v3-Relation) | repeated | object relations |






<a name="aserto-directory-reader-v3-GetObjectsRequest"></a>

### GetObjectsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier (optional) |
| page | [aserto.directory.common.v3.PaginationRequest](#aserto-directory-common-v3-PaginationRequest) |  | pagination request (optional) |






<a name="aserto-directory-reader-v3-GetObjectsResponse"></a>

### GetObjectsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| results | [aserto.directory.common.v3.Object](#aserto-directory-common-v3-Object) | repeated | array of object instances |
| page | [aserto.directory.common.v3.PaginationResponse](#aserto-directory-common-v3-PaginationResponse) |  | pagination response |






<a name="aserto-directory-reader-v3-GetRelationRequest"></a>

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






<a name="aserto-directory-reader-v3-GetRelationResponse"></a>

### GetRelationResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [aserto.directory.common.v3.Relation](#aserto-directory-common-v3-Relation) |  | relation instance |
| objects | [GetRelationResponse.ObjectsEntry](#aserto-directory-reader-v3-GetRelationResponse-ObjectsEntry) | repeated | map of materialized relation objects |






<a name="aserto-directory-reader-v3-GetRelationResponse-ObjectsEntry"></a>

### GetRelationResponse.ObjectsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [aserto.directory.common.v3.Object](#aserto-directory-common-v3-Object) |  |  |






<a name="aserto-directory-reader-v3-GetRelationsRequest"></a>

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
| page | [aserto.directory.common.v3.PaginationRequest](#aserto-directory-common-v3-PaginationRequest) |  | pagination request (optional) |






<a name="aserto-directory-reader-v3-GetRelationsResponse"></a>

### GetRelationsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| results | [aserto.directory.common.v3.Relation](#aserto-directory-common-v3-Relation) | repeated | array of relation instances |
| objects | [GetRelationsResponse.ObjectsEntry](#aserto-directory-reader-v3-GetRelationsResponse-ObjectsEntry) | repeated | map of materialized relation objects |
| page | [aserto.directory.common.v3.PaginationResponse](#aserto-directory-common-v3-PaginationResponse) |  | pagination response |






<a name="aserto-directory-reader-v3-GetRelationsResponse-ObjectsEntry"></a>

### GetRelationsResponse.ObjectsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [aserto.directory.common.v3.Object](#aserto-directory-common-v3-Object) |  |  |





 

 

 


<a name="aserto-directory-reader-v3-Reader"></a>

### Reader


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetObject | [GetObjectRequest](#aserto-directory-reader-v3-GetObjectRequest) | [GetObjectResponse](#aserto-directory-reader-v3-GetObjectResponse) | get object |
| GetObjectMany | [GetObjectManyRequest](#aserto-directory-reader-v3-GetObjectManyRequest) | [GetObjectManyResponse](#aserto-directory-reader-v3-GetObjectManyResponse) | get multiple objects |
| GetObjects | [GetObjectsRequest](#aserto-directory-reader-v3-GetObjectsRequest) | [GetObjectsResponse](#aserto-directory-reader-v3-GetObjectsResponse) | list objects |
| GetRelation | [GetRelationRequest](#aserto-directory-reader-v3-GetRelationRequest) | [GetRelationResponse](#aserto-directory-reader-v3-GetRelationResponse) | get relation |
| GetRelations | [GetRelationsRequest](#aserto-directory-reader-v3-GetRelationsRequest) | [GetRelationsResponse](#aserto-directory-reader-v3-GetRelationsResponse) | list relations |
| Check | [CheckRequest](#aserto-directory-reader-v3-CheckRequest) | [CheckResponse](#aserto-directory-reader-v3-CheckResponse) | check if subject has relation or permission with object |
| Checks | [ChecksRequest](#aserto-directory-reader-v3-ChecksRequest) | [ChecksResponse](#aserto-directory-reader-v3-ChecksResponse) | checks validates a set of check requests in a single roundtrip |
| CheckPermission | [CheckPermissionRequest](#aserto-directory-reader-v3-CheckPermissionRequest) | [CheckPermissionResponse](#aserto-directory-reader-v3-CheckPermissionResponse) | check permission (deprecated, use the check method) Deprecated: use directory.reader.v3.Check() |
| CheckRelation | [CheckRelationRequest](#aserto-directory-reader-v3-CheckRelationRequest) | [CheckRelationResponse](#aserto-directory-reader-v3-CheckRelationResponse) | check relation (deprecated, use the check method) Deprecated: use directory.reader.v3.Check() |
| GetGraph | [GetGraphRequest](#aserto-directory-reader-v3-GetGraphRequest) | [GetGraphResponse](#aserto-directory-reader-v3-GetGraphResponse) | get object relationship graph |

 



<a name="v3_writer_writer-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## v3/writer/writer.proto



<a name="aserto-directory-writer-v3-DeleteObjectRequest"></a>

### DeleteObjectRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |
| with_relations | [bool](#bool) |  | delete object relations, both object and subject relations (optional) |






<a name="aserto-directory-writer-v3-DeleteObjectResponse"></a>

### DeleteObjectResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [google.protobuf.Empty](#google-protobuf-Empty) |  | empty result |






<a name="aserto-directory-writer-v3-DeleteRelationRequest"></a>

### DeleteRelationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |
| relation | [string](#string) |  | object relation name |
| subject_type | [string](#string) |  | subject type identifier |
| subject_id | [string](#string) |  | subject instance identifier |
| subject_relation | [string](#string) |  | subject relation name (optional) |






<a name="aserto-directory-writer-v3-DeleteRelationResponse"></a>

### DeleteRelationResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [google.protobuf.Empty](#google-protobuf-Empty) |  | empty result |






<a name="aserto-directory-writer-v3-SetObjectRequest"></a>

### SetObjectRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object | [aserto.directory.common.v3.Object](#aserto-directory-common-v3-Object) |  | object instance |






<a name="aserto-directory-writer-v3-SetObjectResponse"></a>

### SetObjectResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [aserto.directory.common.v3.Object](#aserto-directory-common-v3-Object) |  | object instance |






<a name="aserto-directory-writer-v3-SetRelationRequest"></a>

### SetRelationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| relation | [aserto.directory.common.v3.Relation](#aserto-directory-common-v3-Relation) |  | relation instance |






<a name="aserto-directory-writer-v3-SetRelationResponse"></a>

### SetRelationResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [aserto.directory.common.v3.Relation](#aserto-directory-common-v3-Relation) |  | relation instance |





 

 

 


<a name="aserto-directory-writer-v3-Writer"></a>

### Writer


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| SetObject | [SetObjectRequest](#aserto-directory-writer-v3-SetObjectRequest) | [SetObjectResponse](#aserto-directory-writer-v3-SetObjectResponse) | set object instance |
| DeleteObject | [DeleteObjectRequest](#aserto-directory-writer-v3-DeleteObjectRequest) | [DeleteObjectResponse](#aserto-directory-writer-v3-DeleteObjectResponse) | delete object instance |
| SetRelation | [SetRelationRequest](#aserto-directory-writer-v3-SetRelationRequest) | [SetRelationResponse](#aserto-directory-writer-v3-SetRelationResponse) | set relation instance |
| DeleteRelation | [DeleteRelationRequest](#aserto-directory-writer-v3-DeleteRelationRequest) | [DeleteRelationResponse](#aserto-directory-writer-v3-DeleteRelationResponse) | delete relation instance |

 



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

