# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [directory/v4/manifest.proto](#directory_v4_manifest-proto)
    - [Manifest](#directory-v4-Manifest)
  
- [directory/v4/model.proto](#directory_v4_model-proto)
    - [Model](#directory-v4-Model)
  
- [directory/v4/object.proto](#directory_v4_object-proto)
    - [Object](#directory-v4-Object)
  
- [directory/v4/object_identifier.proto](#directory_v4_object_identifier-proto)
    - [ObjectIdentifier](#directory-v4-ObjectIdentifier)
  
- [directory/v4/openapi.proto](#directory_v4_openapi-proto)
- [directory/v4/pagination.proto](#directory_v4_pagination-proto)
    - [PaginationRequest](#directory-v4-PaginationRequest)
    - [PaginationResponse](#directory-v4-PaginationResponse)
  
- [directory/v4/reader/check.proto](#directory_v4_reader_check-proto)
    - [CheckRequest](#directory-v4-reader-CheckRequest)
    - [CheckResponse](#directory-v4-reader-CheckResponse)
    - [ChecksRequest](#directory-v4-reader-ChecksRequest)
    - [ChecksResponse](#directory-v4-reader-ChecksResponse)
  
- [directory/v4/relation.proto](#directory_v4_relation-proto)
    - [Relation](#directory-v4-Relation)
  
- [directory/v4/reader/export.proto](#directory_v4_reader_export-proto)
    - [ExportRequest](#directory-v4-reader-ExportRequest)
    - [ExportResponse](#directory-v4-reader-ExportResponse)
  
    - [Option](#directory-v4-reader-Option)
  
- [directory/v4/reader/graph.proto](#directory_v4_reader_graph-proto)
    - [GraphRequest](#directory-v4-reader-GraphRequest)
    - [GraphResponse](#directory-v4-reader-GraphResponse)
  
- [directory/v4/reader/manifest.proto](#directory_v4_reader_manifest-proto)
    - [GetManifestRequest](#directory-v4-reader-GetManifestRequest)
    - [GetManifestResponse](#directory-v4-reader-GetManifestResponse)
  
- [directory/v4/reader/model.proto](#directory_v4_reader_model-proto)
    - [GetModelRequest](#directory-v4-reader-GetModelRequest)
    - [GetModelResponse](#directory-v4-reader-GetModelResponse)
  
- [directory/v4/reader/object.proto](#directory_v4_reader_object-proto)
    - [GetObjectRequest](#directory-v4-reader-GetObjectRequest)
    - [GetObjectResponse](#directory-v4-reader-GetObjectResponse)
    - [ListObjectsRequest](#directory-v4-reader-ListObjectsRequest)
    - [ListObjectsResponse](#directory-v4-reader-ListObjectsResponse)
  
- [directory/v4/reader/relation.proto](#directory_v4_reader_relation-proto)
    - [GetRelationRequest](#directory-v4-reader-GetRelationRequest)
    - [GetRelationResponse](#directory-v4-reader-GetRelationResponse)
    - [GetRelationResponse.ObjectsEntry](#directory-v4-reader-GetRelationResponse-ObjectsEntry)
    - [ListRelationsRequest](#directory-v4-reader-ListRelationsRequest)
    - [ListRelationsResponse](#directory-v4-reader-ListRelationsResponse)
    - [ListRelationsResponse.ObjectsEntry](#directory-v4-reader-ListRelationsResponse-ObjectsEntry)
  
- [directory/v4/reader/service.proto](#directory_v4_reader_service-proto)
    - [Reader](#directory-v4-reader-Reader)
  
- [directory/v4/relation_identifier.proto](#directory_v4_relation_identifier-proto)
    - [RelationIdentifier](#directory-v4-RelationIdentifier)
  
- [directory/v4/writer/import.proto](#directory_v4_writer_import-proto)
    - [ImportCounter](#directory-v4-writer-ImportCounter)
    - [ImportRequest](#directory-v4-writer-ImportRequest)
    - [ImportResponse](#directory-v4-writer-ImportResponse)
    - [ImportStatus](#directory-v4-writer-ImportStatus)
  
    - [Opcode](#directory-v4-writer-Opcode)
  
- [directory/v4/writer/manifest.proto](#directory_v4_writer_manifest-proto)
    - [DeleteManifestRequest](#directory-v4-writer-DeleteManifestRequest)
    - [DeleteManifestResponse](#directory-v4-writer-DeleteManifestResponse)
    - [SetManifestRequest](#directory-v4-writer-SetManifestRequest)
    - [SetManifestResponse](#directory-v4-writer-SetManifestResponse)
  
- [directory/v4/writer/object.proto](#directory_v4_writer_object-proto)
    - [DeleteObjectRequest](#directory-v4-writer-DeleteObjectRequest)
    - [DeleteObjectResponse](#directory-v4-writer-DeleteObjectResponse)
    - [SetObjectRequest](#directory-v4-writer-SetObjectRequest)
    - [SetObjectResponse](#directory-v4-writer-SetObjectResponse)
  
- [directory/v4/writer/relation.proto](#directory_v4_writer_relation-proto)
    - [DeleteRelationRequest](#directory-v4-writer-DeleteRelationRequest)
    - [DeleteRelationResponse](#directory-v4-writer-DeleteRelationResponse)
    - [SetRelationRequest](#directory-v4-writer-SetRelationRequest)
    - [SetRelationResponse](#directory-v4-writer-SetRelationResponse)
  
- [directory/v4/writer/service.proto](#directory_v4_writer_service-proto)
    - [Writer](#directory-v4-writer-Writer)
  
- [Scalar Value Types](#scalar-value-types)



<a name="directory_v4_manifest-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## directory/v4/manifest.proto



<a name="directory-v4-Manifest"></a>

### Manifest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  | manifest identifier |
| content | [bytes](#bytes) |  | manifest content data |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | last updated timestamp (UTC) |
| etag | [string](#string) |  | object instance etag (optional) |





 

 

 

 



<a name="directory_v4_model-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## directory/v4/model.proto



<a name="directory-v4-Model"></a>

### Model



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  | model identifier |
| model | [google.protobuf.Struct](#google-protobuf-Struct) |  | model |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | last updated timestamp (UTC) |
| etag | [string](#string) |  | object instance etag (optional) |





 

 

 

 



<a name="directory_v4_object-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## directory/v4/object.proto



<a name="directory-v4-Object"></a>

### Object



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |
| properties | [google.protobuf.Struct](#google-protobuf-Struct) |  | property bag (optional) |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | last updated timestamp (UTC) |
| etag | [string](#string) |  | object instance etag (optional) |





 

 

 

 



<a name="directory_v4_object_identifier-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## directory/v4/object_identifier.proto



<a name="directory-v4-ObjectIdentifier"></a>

### ObjectIdentifier
Object identifier


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |





 

 

 

 



<a name="directory_v4_openapi-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## directory/v4/openapi.proto


 

 

 

 



<a name="directory_v4_pagination-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## directory/v4/pagination.proto



<a name="directory-v4-PaginationRequest"></a>

### PaginationRequest
Pagination request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size | [int32](#int32) |  | requested page size, valid value between 1-100 rows (optional, default 100) |
| token | [string](#string) |  | pagination start token (optional, default &#34;&#34;) |






<a name="directory-v4-PaginationResponse"></a>

### PaginationResponse
Pagination response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| next_token | [string](#string) |  | next page token, when empty there are no more pages to fetch |





 

 

 

 



<a name="directory_v4_reader_check-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## directory/v4/reader/check.proto



<a name="directory-v4-reader-CheckRequest"></a>

### CheckRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |
| relation | [string](#string) |  | relation name |
| subject_type | [string](#string) |  | subject type identifier |
| subject_id | [string](#string) |  | subject instance identifier |
| trace | [bool](#bool) |  | collect trace information (optional) |






<a name="directory-v4-reader-CheckResponse"></a>

### CheckResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| check | [bool](#bool) |  | check result |
| trace | [string](#string) | repeated | trace information |
| context | [google.protobuf.Struct](#google-protobuf-Struct) |  | context |






<a name="directory-v4-reader-ChecksRequest"></a>

### ChecksRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| default | [CheckRequest](#directory-v4-reader-CheckRequest) |  |  |
| checks | [CheckRequest](#directory-v4-reader-CheckRequest) | repeated |  |






<a name="directory-v4-reader-ChecksResponse"></a>

### ChecksResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| checks | [CheckResponse](#directory-v4-reader-CheckResponse) | repeated |  |





 

 

 

 



<a name="directory_v4_relation-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## directory/v4/relation.proto



<a name="directory-v4-Relation"></a>

### Relation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |
| relation | [string](#string) |  | object relation name |
| subject_type | [string](#string) |  | subject type identifier |
| subject_id | [string](#string) |  | subject instance identifier |
| subject_relation | [string](#string) |  | subject relation name (optional) |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | last updated timestamp (UTC) |
| etag | [string](#string) |  | object instance etag (optional) |





 

 

 

 



<a name="directory_v4_reader_export-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## directory/v4/reader/export.proto



<a name="directory-v4-reader-ExportRequest"></a>

### ExportRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| options | [uint32](#uint32) |  | data export options mask |
| start_from | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | start export from timestamp (UTC) |






<a name="directory-v4-reader-ExportResponse"></a>

### ExportResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object | [directory.v4.Object](#directory-v4-Object) |  | object instance (data) |
| relation | [directory.v4.Relation](#directory-v4-Relation) |  | relation instance (data) |
| stats | [google.protobuf.Struct](#google-protobuf-Struct) |  | object and/or relation stats (no data) |





 


<a name="directory-v4-reader-Option"></a>

### Option


| Name | Number | Description |
| ---- | ------ | ----------- |
| OPTION_UNKNOWN | 0 | nothing selected (default initialization value) |
| OPTION_DATA_OBJECTS | 8 | object instances |
| OPTION_DATA_RELATIONS | 16 | relation instances |
| OPTION_DATA | 24 | all data = OPTION_DATA_OBJECTS | OPTION_DATA_RELATIONS |
| OPTION_STATS | 64 | stats |


 

 

 



<a name="directory_v4_reader_graph-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## directory/v4/reader/graph.proto



<a name="directory-v4-reader-GraphRequest"></a>

### GraphRequest



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






<a name="directory-v4-reader-GraphResponse"></a>

### GraphResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| results | [directory.v4.ObjectIdentifier](#directory-v4-ObjectIdentifier) | repeated | matching object identifiers |
| explanation | [google.protobuf.Struct](#google-protobuf-Struct) |  | explanation of results |
| trace | [string](#string) | repeated | trace information |





 

 

 

 



<a name="directory_v4_reader_manifest-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## directory/v4/reader/manifest.proto



<a name="directory-v4-reader-GetManifestRequest"></a>

### GetManifestRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="directory-v4-reader-GetManifestResponse"></a>

### GetManifestResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| manifest | [directory.v4.Manifest](#directory-v4-Manifest) |  |  |





 

 

 

 



<a name="directory_v4_reader_model-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## directory/v4/reader/model.proto



<a name="directory-v4-reader-GetModelRequest"></a>

### GetModelRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="directory-v4-reader-GetModelResponse"></a>

### GetModelResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| model | [directory.v4.Model](#directory-v4-Model) |  |  |





 

 

 

 



<a name="directory_v4_reader_object-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## directory/v4/reader/object.proto



<a name="directory-v4-reader-GetObjectRequest"></a>

### GetObjectRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |
| with_relations | [bool](#bool) |  | materialize the object relations objects (optional) |






<a name="directory-v4-reader-GetObjectResponse"></a>

### GetObjectResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [directory.v4.Object](#directory-v4-Object) |  | object instance |
| relations | [directory.v4.Relation](#directory-v4-Relation) | repeated | object relations |






<a name="directory-v4-reader-ListObjectsRequest"></a>

### ListObjectsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier (optional) |
| page | [directory.v4.PaginationRequest](#directory-v4-PaginationRequest) |  | pagination request (optional) |






<a name="directory-v4-reader-ListObjectsResponse"></a>

### ListObjectsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| results | [directory.v4.Object](#directory-v4-Object) | repeated | array of object instances |
| page | [directory.v4.PaginationResponse](#directory-v4-PaginationResponse) |  | pagination response |





 

 

 

 



<a name="directory_v4_reader_relation-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## directory/v4/reader/relation.proto



<a name="directory-v4-reader-GetRelationRequest"></a>

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






<a name="directory-v4-reader-GetRelationResponse"></a>

### GetRelationResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [directory.v4.Relation](#directory-v4-Relation) |  | relation instance |
| objects | [GetRelationResponse.ObjectsEntry](#directory-v4-reader-GetRelationResponse-ObjectsEntry) | repeated | map of materialized relation objects |






<a name="directory-v4-reader-GetRelationResponse-ObjectsEntry"></a>

### GetRelationResponse.ObjectsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [directory.v4.Object](#directory-v4-Object) |  |  |






<a name="directory-v4-reader-ListRelationsRequest"></a>

### ListRelationsRequest



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
| page | [directory.v4.PaginationRequest](#directory-v4-PaginationRequest) |  | pagination request (optional) |






<a name="directory-v4-reader-ListRelationsResponse"></a>

### ListRelationsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| results | [directory.v4.Relation](#directory-v4-Relation) | repeated | array of relation instances |
| objects | [ListRelationsResponse.ObjectsEntry](#directory-v4-reader-ListRelationsResponse-ObjectsEntry) | repeated | map of materialized relation objects |
| page | [directory.v4.PaginationResponse](#directory-v4-PaginationResponse) |  | pagination response |






<a name="directory-v4-reader-ListRelationsResponse-ObjectsEntry"></a>

### ListRelationsResponse.ObjectsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [directory.v4.Object](#directory-v4-Object) |  |  |





 

 

 

 



<a name="directory_v4_reader_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## directory/v4/reader/service.proto


 

 

 


<a name="directory-v4-reader-Reader"></a>

### Reader


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetManifest | [GetManifestRequest](#directory-v4-reader-GetManifestRequest) | [GetManifestResponse](#directory-v4-reader-GetManifestResponse) | get manifest |
| GetModel | [GetModelRequest](#directory-v4-reader-GetModelRequest) | [GetModelResponse](#directory-v4-reader-GetModelResponse) | get model |
| GetObject | [GetObjectRequest](#directory-v4-reader-GetObjectRequest) | [GetObjectResponse](#directory-v4-reader-GetObjectResponse) | get object |
| ListObjects | [ListObjectsRequest](#directory-v4-reader-ListObjectsRequest) | [ListObjectsResponse](#directory-v4-reader-ListObjectsResponse) | list objects |
| GetRelation | [GetRelationRequest](#directory-v4-reader-GetRelationRequest) | [GetRelationResponse](#directory-v4-reader-GetRelationResponse) | get relation |
| ListRelations | [ListRelationsRequest](#directory-v4-reader-ListRelationsRequest) | [ListRelationsResponse](#directory-v4-reader-ListRelationsResponse) | list relations |
| Check | [CheckRequest](#directory-v4-reader-CheckRequest) | [CheckResponse](#directory-v4-reader-CheckResponse) | check if subject has relation or permission with object |
| Checks | [ChecksRequest](#directory-v4-reader-ChecksRequest) | [ChecksResponse](#directory-v4-reader-ChecksResponse) | checks validates a set of check requests in a single roundtrip |
| Graph | [GraphRequest](#directory-v4-reader-GraphRequest) | [GraphResponse](#directory-v4-reader-GraphResponse) | get object relationship graph |
| Export | [ExportRequest](#directory-v4-reader-ExportRequest) | [ExportResponse](#directory-v4-reader-ExportResponse) stream | export objects and relations as a stream |

 



<a name="directory_v4_relation_identifier-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## directory/v4/relation_identifier.proto



<a name="directory-v4-RelationIdentifier"></a>

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





 

 

 

 



<a name="directory_v4_writer_import-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## directory/v4/writer/import.proto



<a name="directory-v4-writer-ImportCounter"></a>

### ImportCounter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| recv | [uint64](#uint64) |  | number of messages received |
| set | [uint64](#uint64) |  | number of messages with OPCODE_SET |
| delete | [uint64](#uint64) |  | number of messages with OPCODE_DELETE |
| error | [uint64](#uint64) |  | number of messages resulting in error |
| type | [string](#string) |  | counter of type (object|relation) |






<a name="directory-v4-writer-ImportRequest"></a>

### ImportRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| op_code | [Opcode](#directory-v4-writer-Opcode) |  | operation Opcode enum value |
| object | [directory.v4.Object](#directory-v4-Object) |  | object import message |
| relation | [directory.v4.Relation](#directory-v4-Relation) |  | relation import message |






<a name="directory-v4-writer-ImportResponse"></a>

### ImportResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object | [ImportCounter](#directory-v4-writer-ImportCounter) |  | **Deprecated.** object import counter |
| relation | [ImportCounter](#directory-v4-writer-ImportCounter) |  | **Deprecated.** relation import counter |
| status | [ImportStatus](#directory-v4-writer-ImportStatus) |  | import status message |
| counter | [ImportCounter](#directory-v4-writer-ImportCounter) |  | import counter per type |






<a name="directory-v4-writer-ImportStatus"></a>

### ImportStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [uint32](#uint32) |  | gRPC status code (google.golang.org/grpc/codes) |
| msg | [string](#string) |  | gRPC status message (google.golang.org/grpc/status) |
| req | [ImportRequest](#directory-v4-writer-ImportRequest) |  | req contains the original import request message |





 


<a name="directory-v4-writer-Opcode"></a>

### Opcode


| Name | Number | Description |
| ---- | ------ | ----------- |
| OPCODE_UNKNOWN | 0 |  |
| OPCODE_SET | 1 |  |
| OPCODE_DELETE | 2 |  |
| OPCODE_DELETE_WITH_RELATIONS | 3 |  |


 

 

 



<a name="directory_v4_writer_manifest-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## directory/v4/writer/manifest.proto



<a name="directory-v4-writer-DeleteManifestRequest"></a>

### DeleteManifestRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="directory-v4-writer-DeleteManifestResponse"></a>

### DeleteManifestResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [google.protobuf.Empty](#google-protobuf-Empty) |  | empty result |






<a name="directory-v4-writer-SetManifestRequest"></a>

### SetManifestRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| manifest | [directory.v4.Manifest](#directory-v4-Manifest) |  |  |






<a name="directory-v4-writer-SetManifestResponse"></a>

### SetManifestResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| manifest | [directory.v4.Manifest](#directory-v4-Manifest) |  |  |





 

 

 

 



<a name="directory_v4_writer_object-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## directory/v4/writer/object.proto



<a name="directory-v4-writer-DeleteObjectRequest"></a>

### DeleteObjectRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |
| with_relations | [bool](#bool) |  | delete object relations, both object and subject relations (optional) |






<a name="directory-v4-writer-DeleteObjectResponse"></a>

### DeleteObjectResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [google.protobuf.Empty](#google-protobuf-Empty) |  | empty result |






<a name="directory-v4-writer-SetObjectRequest"></a>

### SetObjectRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object | [directory.v4.Object](#directory-v4-Object) |  | object instance |






<a name="directory-v4-writer-SetObjectResponse"></a>

### SetObjectResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [directory.v4.Object](#directory-v4-Object) |  | object instance |





 

 

 

 



<a name="directory_v4_writer_relation-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## directory/v4/writer/relation.proto



<a name="directory-v4-writer-DeleteRelationRequest"></a>

### DeleteRelationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_type | [string](#string) |  | object type identifier |
| object_id | [string](#string) |  | object instance identifier |
| relation | [string](#string) |  | object relation name |
| subject_type | [string](#string) |  | subject type identifier |
| subject_id | [string](#string) |  | subject instance identifier |
| subject_relation | [string](#string) |  | subject relation name (optional) |






<a name="directory-v4-writer-DeleteRelationResponse"></a>

### DeleteRelationResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [google.protobuf.Empty](#google-protobuf-Empty) |  | empty result |






<a name="directory-v4-writer-SetRelationRequest"></a>

### SetRelationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| relation | [directory.v4.Relation](#directory-v4-Relation) |  | relation instance |






<a name="directory-v4-writer-SetRelationResponse"></a>

### SetRelationResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [directory.v4.Relation](#directory-v4-Relation) |  | relation instance |





 

 

 

 



<a name="directory_v4_writer_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## directory/v4/writer/service.proto


 

 

 


<a name="directory-v4-writer-Writer"></a>

### Writer


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| SetManifest | [SetManifestRequest](#directory-v4-writer-SetManifestRequest) | [SetManifestResponse](#directory-v4-writer-SetManifestResponse) | set manifest |
| DeleteManifest | [DeleteManifestRequest](#directory-v4-writer-DeleteManifestRequest) | [DeleteManifestResponse](#directory-v4-writer-DeleteManifestResponse) | delete manifest |
| SetObject | [SetObjectRequest](#directory-v4-writer-SetObjectRequest) | [SetObjectResponse](#directory-v4-writer-SetObjectResponse) | set object instance |
| DeleteObject | [DeleteObjectRequest](#directory-v4-writer-DeleteObjectRequest) | [DeleteObjectResponse](#directory-v4-writer-DeleteObjectResponse) | delete object instance |
| SetRelation | [SetRelationRequest](#directory-v4-writer-SetRelationRequest) | [SetRelationResponse](#directory-v4-writer-SetRelationResponse) | set relation instance |
| DeleteRelation | [DeleteRelationRequest](#directory-v4-writer-DeleteRelationRequest) | [DeleteRelationResponse](#directory-v4-writer-DeleteRelationResponse) | delete relation instance |
| Import | [ImportRequest](#directory-v4-writer-ImportRequest) stream | [ImportResponse](#directory-v4-writer-ImportResponse) stream | import stream of objects and relations |

 



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

