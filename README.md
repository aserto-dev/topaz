# TOPAZ

Welcome to Topaz!

Topaz presents an opinionated Open Source Authorization solution that combines policy-as-code with policy-as-data. It provides an end-to-end solution for application authorization that is buit on top of the Open Policy Agent (OPA) and influenced by Google Zanzibar.

With Topaz you can scale your authorization model from RBAC to ABAC and ReBAC, while retaining the benefits of policy-as-code, decision logging, and a local deployment model.

## Capabilites 
1. Open-source authorizer delivered as a container and deployed next to your app - evaluating authorization decisions in milliseconds, at 100% availability
2. First-class support for relationship-based access control (ReBAC) via a local database and a set of built-ins
3. Define authorization policy-as-code, and build it into an immutable, signed OCI image
4. gRPC, REST APIs, and language bindings for evaluating permissions, as well as storing user, resource, and relationship data right next to the decision engine
5. Capture detailed decision logs for every authorization decision

## Built upon an open foundation
1. OPA-based open-source authorizer, purpose-built for API and application authorization
2. Out of the box support for RBAC, ABAC, and ReBAC authorization models
3. Integrated Zanzibar-based directory for evaluating relationship-based decisions
