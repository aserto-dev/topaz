package openapi

// Expose some unexported functions for unit testing.
var (
	Filter    = filter
	ParseSpec = parseSpec
	MatchAny  = matchAny
)
