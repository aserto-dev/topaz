# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [v2/api/api.proto](#v2_api_api-proto)
    - [Decision](#aserto-authorizer-v2-api-Decision)
    - [Decision.AnnotationsEntry](#aserto-authorizer-v2-api-Decision-AnnotationsEntry)
    - [Decision.OutcomesEntry](#aserto-authorizer-v2-api-Decision-OutcomesEntry)
    - [DecisionPolicy](#aserto-authorizer-v2-api-DecisionPolicy)
    - [DecisionUser](#aserto-authorizer-v2-api-DecisionUser)
    - [IdentityContext](#aserto-authorizer-v2-api-IdentityContext)
    - [Module](#aserto-authorizer-v2-api-Module)
    - [PolicyContext](#aserto-authorizer-v2-api-PolicyContext)
    - [PolicyInstance](#aserto-authorizer-v2-api-PolicyInstance)
  
    - [IdentityType](#aserto-authorizer-v2-api-IdentityType)
  
- [v2/authorizer.proto](#v2_authorizer-proto)
    - [CompileRequest](#aserto-authorizer-v2-CompileRequest)
    - [CompileResponse](#aserto-authorizer-v2-CompileResponse)
    - [Decision](#aserto-authorizer-v2-Decision)
    - [DecisionTreeOptions](#aserto-authorizer-v2-DecisionTreeOptions)
    - [DecisionTreeRequest](#aserto-authorizer-v2-DecisionTreeRequest)
    - [DecisionTreeResponse](#aserto-authorizer-v2-DecisionTreeResponse)
    - [GetPolicyRequest](#aserto-authorizer-v2-GetPolicyRequest)
    - [GetPolicyResponse](#aserto-authorizer-v2-GetPolicyResponse)
    - [InfoRequest](#aserto-authorizer-v2-InfoRequest)
    - [InfoResponse](#aserto-authorizer-v2-InfoResponse)
    - [IsRequest](#aserto-authorizer-v2-IsRequest)
    - [IsResponse](#aserto-authorizer-v2-IsResponse)
    - [ListPoliciesRequest](#aserto-authorizer-v2-ListPoliciesRequest)
    - [ListPoliciesResponse](#aserto-authorizer-v2-ListPoliciesResponse)
    - [QueryOptions](#aserto-authorizer-v2-QueryOptions)
    - [QueryRequest](#aserto-authorizer-v2-QueryRequest)
    - [QueryResponse](#aserto-authorizer-v2-QueryResponse)
  
    - [PathSeparator](#aserto-authorizer-v2-PathSeparator)
    - [TraceLevel](#aserto-authorizer-v2-TraceLevel)
  
    - [Authorizer](#aserto-authorizer-v2-Authorizer)
  
- [Scalar Value Types](#scalar-value-types)



<a name="v2_api_api-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## v2/api/api.proto



<a name="aserto-authorizer-v2-api-Decision"></a>

### Decision
represents a decision that an authorizer performed in the past


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  | id, unique id, replay a decision starting with this, also useful to de-dup |
| timestamp | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | timestamp, UTC time when the decision was made |
| path | [string](#string) |  | path, policy path used in decision |
| user | [DecisionUser](#aserto-authorizer-v2-api-DecisionUser) |  | user, info about user for whom the decision as made |
| policy | [DecisionPolicy](#aserto-authorizer-v2-api-DecisionPolicy) |  | policy, info about policy used for the decision |
| outcomes | [Decision.OutcomesEntry](#aserto-authorizer-v2-api-Decision-OutcomesEntry) | repeated | outcomes of the decisions specified in the policy context |
| resource | [google.protobuf.Struct](#google-protobuf-Struct) |  | resource context used by the decision |
| annotations | [Decision.AnnotationsEntry](#aserto-authorizer-v2-api-Decision-AnnotationsEntry) | repeated | annotations that may be added to a decision |






<a name="aserto-authorizer-v2-api-Decision-AnnotationsEntry"></a>

### Decision.AnnotationsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="aserto-authorizer-v2-api-Decision-OutcomesEntry"></a>

### Decision.OutcomesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [bool](#bool) |  |  |






<a name="aserto-authorizer-v2-api-DecisionPolicy"></a>

### DecisionPolicy
information about a policy used in a decision


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| context | [PolicyContext](#aserto-authorizer-v2-api-PolicyContext) |  | policy context used in the decision |
| registry_service | [string](#string) |  | registry service where policy was retrieved from (e.g. ghcr.io) |
| registry_image | [string](#string) |  | registry image of the policy in the registry, including org (e.g. acmecorp/peoplefinder-abac) |
| registry_tag | [string](#string) |  | registry tag of the policy image (e.g. 0.8.2 or latest) |
| registry_digest | [string](#string) |  | registry digest of the policy image |
| policy_instance | [PolicyInstance](#aserto-authorizer-v2-api-PolicyInstance) | optional | **Deprecated.** policy instance used in decision |






<a name="aserto-authorizer-v2-api-DecisionUser"></a>

### DecisionUser
information about a user on behalf of whom a decision was made


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| context | [IdentityContext](#aserto-authorizer-v2-api-IdentityContext) |  | identity context used in the decision |
| id | [string](#string) |  | id of the user the identity resolved to |
| email | [string](#string) |  | convenience human-readable identifier |






<a name="aserto-authorizer-v2-api-IdentityContext"></a>

### IdentityContext



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| identity | [string](#string) |  | identity |
| type | [IdentityType](#aserto-authorizer-v2-api-IdentityType) |  | identity type |






<a name="aserto-authorizer-v2-api-Module"></a>

### Module
module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) | optional | id |
| raw | [string](#string) | optional | raw |
| package_path | [string](#string) | optional | package path |
| ast | [google.protobuf.Value](#google-protobuf-Value) | optional | ast |
| package_root | [string](#string) | optional | package root |






<a name="aserto-authorizer-v2-api-PolicyContext"></a>

### PolicyContext



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [string](#string) |  | policy path aka package name |
| decisions | [string](#string) | repeated | decisions list (1..N) of policy decisions (aka rules) |






<a name="aserto-authorizer-v2-api-PolicyInstance"></a>

### PolicyInstance
Deprecated: No longer required.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | **Deprecated.** policy name |
| instance_label | [string](#string) |  | **Deprecated.** label identifying the instance of the policy |





 


<a name="aserto-authorizer-v2-api-IdentityType"></a>

### IdentityType
Identity types, describes the payload type of the identity field inside the IdentityContext message.

| Name | Number | Description |
| ---- | ------ | ----------- |
| IDENTITY_TYPE_UNKNOWN | 0 | Unknown, value not set, requests will fail with identity type not set error. |
| IDENTITY_TYPE_NONE | 1 | None, no explicit identity context set, equals anonymous. |
| IDENTITY_TYPE_SUB | 2 | Sub(ject), identity field contains an oAUTH subject. |
| IDENTITY_TYPE_JWT | 3 | JWT, identity field contains a JWT access token. |
| IDENTITY_TYPE_MANUAL | 4 | Manual, propagates thw identity field as-is, without validation, into the input object. |


 

 

 



<a name="v2_authorizer-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## v2/authorizer.proto



<a name="aserto-authorizer-v2-CompileRequest"></a>

### CompileRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| query | [string](#string) |  | query |
| input | [string](#string) |  | input block |
| unknowns | [string](#string) | repeated | unknowns |
| disable_inlining | [string](#string) | repeated | disable inlining |
| options | [QueryOptions](#aserto-authorizer-v2-QueryOptions) | optional | query options |
| policy_context | [api.PolicyContext](#aserto-authorizer-v2-api-PolicyContext) | optional | policy context |
| identity_context | [api.IdentityContext](#aserto-authorizer-v2-api-IdentityContext) | optional | identity context |
| resource_context | [google.protobuf.Struct](#google-protobuf-Struct) | optional | resource context |
| policy_instance | [api.PolicyInstance](#aserto-authorizer-v2-api-PolicyInstance) | optional | **Deprecated.** policy instance |






<a name="aserto-authorizer-v2-CompileResponse"></a>

### CompileResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |
| metrics | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |
| trace | [google.protobuf.Struct](#google-protobuf-Struct) | repeated |  |
| trace_summary | [string](#string) | repeated |  |






<a name="aserto-authorizer-v2-Decision"></a>

### Decision



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| decision | [string](#string) |  | decision |
| is | [bool](#bool) |  | is (decision outcome) |






<a name="aserto-authorizer-v2-DecisionTreeOptions"></a>

### DecisionTreeOptions



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path_separator | [PathSeparator](#aserto-authorizer-v2-PathSeparator) |  |  |






<a name="aserto-authorizer-v2-DecisionTreeRequest"></a>

### DecisionTreeRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| policy_context | [api.PolicyContext](#aserto-authorizer-v2-api-PolicyContext) |  | policy context |
| identity_context | [api.IdentityContext](#aserto-authorizer-v2-api-IdentityContext) |  | identity context |
| options | [DecisionTreeOptions](#aserto-authorizer-v2-DecisionTreeOptions) |  | decision tree options |
| resource_context | [google.protobuf.Struct](#google-protobuf-Struct) |  | resource context |
| policy_instance | [api.PolicyInstance](#aserto-authorizer-v2-api-PolicyInstance) | optional | **Deprecated.** policy instance |






<a name="aserto-authorizer-v2-DecisionTreeResponse"></a>

### DecisionTreeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path_root | [string](#string) |  | path root |
| path | [google.protobuf.Struct](#google-protobuf-Struct) |  | path |






<a name="aserto-authorizer-v2-GetPolicyRequest"></a>

### GetPolicyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| field_mask | [google.protobuf.FieldMask](#google-protobuf-FieldMask) |  |  |
| policy_instance | [api.PolicyInstance](#aserto-authorizer-v2-api-PolicyInstance) | optional | **Deprecated.**  |






<a name="aserto-authorizer-v2-GetPolicyResponse"></a>

### GetPolicyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [api.Module](#aserto-authorizer-v2-api-Module) |  |  |






<a name="aserto-authorizer-v2-InfoRequest"></a>

### InfoRequest







<a name="aserto-authorizer-v2-InfoResponse"></a>

### InfoResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| version | [string](#string) |  |  |
| commit | [string](#string) |  |  |
| date | [string](#string) |  |  |
| os | [string](#string) |  |  |
| arch | [string](#string) |  |  |






<a name="aserto-authorizer-v2-IsRequest"></a>

### IsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| policy_context | [api.PolicyContext](#aserto-authorizer-v2-api-PolicyContext) |  | policy context |
| identity_context | [api.IdentityContext](#aserto-authorizer-v2-api-IdentityContext) |  | identity context |
| resource_context | [google.protobuf.Struct](#google-protobuf-Struct) |  | resource context |
| policy_instance | [api.PolicyInstance](#aserto-authorizer-v2-api-PolicyInstance) | optional | **Deprecated.** policy instance |






<a name="aserto-authorizer-v2-IsResponse"></a>

### IsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| decisions | [Decision](#aserto-authorizer-v2-Decision) | repeated | decisions |






<a name="aserto-authorizer-v2-ListPoliciesRequest"></a>

### ListPoliciesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| field_mask | [google.protobuf.FieldMask](#google-protobuf-FieldMask) |  |  |
| policy_instance | [api.PolicyInstance](#aserto-authorizer-v2-api-PolicyInstance) | optional | **Deprecated.**  |






<a name="aserto-authorizer-v2-ListPoliciesResponse"></a>

### ListPoliciesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| result | [api.Module](#aserto-authorizer-v2-api-Module) | repeated |  |






<a name="aserto-authorizer-v2-QueryOptions"></a>

### QueryOptions



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metrics | [bool](#bool) |  | metrics (default false) |
| instrument | [bool](#bool) |  | instrument (default false)

default false |
| trace | [TraceLevel](#aserto-authorizer-v2-TraceLevel) |  | trace level (default ExplainOffV1) |
| trace_summary | [bool](#bool) |  | trace summary (default false) |






<a name="aserto-authorizer-v2-QueryRequest"></a>

### QueryRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| query | [string](#string) |  | query |
| input | [string](#string) |  | query input block |
| options | [QueryOptions](#aserto-authorizer-v2-QueryOptions) | optional | query options |
| policy_context | [api.PolicyContext](#aserto-authorizer-v2-api-PolicyContext) | optional | policy context |
| identity_context | [api.IdentityContext](#aserto-authorizer-v2-api-IdentityContext) | optional | identity context |
| resource_context | [google.protobuf.Struct](#google-protobuf-Struct) | optional | resource context |
| policy_instance | [api.PolicyInstance](#aserto-authorizer-v2-api-PolicyInstance) | optional | **Deprecated.** policy instance |






<a name="aserto-authorizer-v2-QueryResponse"></a>

### QueryResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| response | [google.protobuf.Struct](#google-protobuf-Struct) |  | query response |
| metrics | [google.protobuf.Struct](#google-protobuf-Struct) |  | metrics |
| trace | [google.protobuf.Struct](#google-protobuf-Struct) | repeated | trace |
| trace_summary | [string](#string) | repeated | trace summary |





 


<a name="aserto-authorizer-v2-PathSeparator"></a>

### PathSeparator


| Name | Number | Description |
| ---- | ------ | ----------- |
| PATH_SEPARATOR_UNKNOWN | 0 | Value not set. |
| PATH_SEPARATOR_DOT | 1 | Dot &#34;.&#34; path separator |
| PATH_SEPARATOR_SLASH | 2 | Slash &#34;/&#34; path separator |



<a name="aserto-authorizer-v2-TraceLevel"></a>

### TraceLevel


| Name | Number | Description |
| ---- | ------ | ----------- |
| TRACE_LEVEL_UNKNOWN | 0 | Value not set. |
| TRACE_LEVEL_OFF | 1 | ExplainOffV1 ExplainModeV1 = &#34;off&#34; |
| TRACE_LEVEL_FULL | 2 | ExplainFullV1 ExplainModeV1 = &#34;full&#34; |
| TRACE_LEVEL_NOTES | 3 | ExplainNotesV1 ExplainModeV1 = &#34;notes&#34; |
| TRACE_LEVEL_FAILS | 4 | ExplainFailsV1 ExplainModeV1 = &#34;fails&#34; |


 

 


<a name="aserto-authorizer-v2-Authorizer"></a>

### Authorizer


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| DecisionTree | [DecisionTreeRequest](#aserto-authorizer-v2-DecisionTreeRequest) | [DecisionTreeResponse](#aserto-authorizer-v2-DecisionTreeResponse) |  |
| Is | [IsRequest](#aserto-authorizer-v2-IsRequest) | [IsResponse](#aserto-authorizer-v2-IsResponse) |  |
| Query | [QueryRequest](#aserto-authorizer-v2-QueryRequest) | [QueryResponse](#aserto-authorizer-v2-QueryResponse) |  |
| Compile | [CompileRequest](#aserto-authorizer-v2-CompileRequest) | [CompileResponse](#aserto-authorizer-v2-CompileResponse) |  |
| ListPolicies | [ListPoliciesRequest](#aserto-authorizer-v2-ListPoliciesRequest) | [ListPoliciesResponse](#aserto-authorizer-v2-ListPoliciesResponse) |  |
| GetPolicy | [GetPolicyRequest](#aserto-authorizer-v2-GetPolicyRequest) | [GetPolicyResponse](#aserto-authorizer-v2-GetPolicyResponse) |  |
| Info | [InfoRequest](#aserto-authorizer-v2-InfoRequest) | [InfoResponse](#aserto-authorizer-v2-InfoResponse) |  |

 



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

