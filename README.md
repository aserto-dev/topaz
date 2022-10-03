# TOPAZ

Welcome to Topaz: Open-source authorization!

Topaz is an open-source authorization service providing fine-grained, real-time, policy-based access control for applications and APIs. 

## Purpose-built fine-grained authorization

1. Fine-grained access control is a critical part of every production-ready application. Topaz is a purpose-built fine-grained access control service that is easy to integrate into your existing stack
2. Start with a fine-grained authorization model that grows with your application: use RBAC, ABAC and ReBAC interchangeably
3. Combine “policy-as-code” and “policy-as-data” to enjoy the best of both worlds and complete flexibility
4. Maintain a clear separation of concerns - extract authorization policy from your application code and into the hands of AppSec
5. Define you authorization policy as code, in Rego, and build it into an immutable, signed OCI image
6. Topaz makes it easy to bring user and resource data close to your authorizer via a local database and set of built-ins


## Built upon an open foundation
1. OPA-based open-source authorizer, purpose-built for API and application authorization
2. Out of the box support for RBAC, ABAC, and ReBAC authorization models
3. Integrated Zanzibar-based directory for evaluating relationship-based decisions


## Benefits 
1. The Authorize lives close to your application. Deploy the authorizer either as a sidecar or a microservice and maintain low latency and high availability.  
2. Bring user and resource data to the authorizer using easy-to-use APIs and a counterpart CLI. This ensures data used by the authorizer is also localized - which is critical for the decisions to be made as quickly as possible. 
3. Combine “Policy-as-Code” and “Policy-as-Data” to build fine-grained authorization models 
4. Consume decision logs produced by your edge authorizer and process them with your favorite analytics platform
5. Brings the best of library and a service - ensure highly performant authorization while keeping your authorization logic separate from your code.

## How to use Topaz 
1. Define your domain model 
2. Load your data
3. Write your policy
4. Deploy the Authorizer
5. Use the Topaz SDKs in your application to make authorization decisions
6. Keep the policy and data up to date
