/*
Package client

The aserto package provides access to the Aserto authorizer and supporting service.

Authorization requests are performed using an AuthorizerClient.
A client can be used on its own to make authorization calls or, more commonly, it can be used to create server
middleware.

# AuthorizerClient

The AuthorizerClient interface, defined in "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2",
describes the operations exposed by the Aserto authorizer service.

Two implementation of AuthorizerClient are available:

1. `authorizer/grpc` provides a client that communicates with the authorizer using gRPC.

2. `authorizer/http` provides a client that communicates with the authorizer over its REST HTTP endpoints.

# Middleware

Two middleware implementations are available in subpackages:

1. middleware/grpc provides middleware for gRPC servers.

2. middleware/http provides middleware for HTTP REST servers.

When authorization middleware is configured and attached to a server, it examines incoming requests, extracts
authorization parameters like the caller's identity, calls the Aserto authorizers, and rejects messages if their
access is denied.

# Other Services

In addition to the authorizer service, go-aserto provides gRPC clients for Aserto's administrative services,
allowing users to programmatically manage their aserto account.

There are two top-level services, each with its own set of sub-services.

1. `client/authorizer` defines a client for services run at the edge and used to serve authorization requests.
*/
package client
