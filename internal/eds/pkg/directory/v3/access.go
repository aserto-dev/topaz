package v3

import (
	"context"

	"github.com/aserto-dev/azm/cache"
	dsr "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dsa "github.com/authzen/access.go/api/access/v1"
	"github.com/rs/zerolog"
)

type Access struct {
	dsa.UnimplementedAccessServer

	logger *zerolog.Logger
	reader *Reader
}

func NewAccess(logger *zerolog.Logger, reader *Reader) *Access {
	return &Access{
		logger: logger,
		reader: reader,
	}
}

// Evaluation access check.
//
// The Access Evaluation API defines the message exchange pattern between a client (PEP)
// and an authorization service (PDP) for executing a single access evaluation.
func (s *Access) Evaluation(ctx context.Context, req *dsa.EvaluationRequest) (*dsa.EvaluationResponse, error) {
	resp, err := s.reader.Check(ctx, extractCheck(req))
	if err != nil {
		return &dsa.EvaluationResponse{}, err
	}

	return &dsa.EvaluationResponse{
		Decision: resp.GetCheck(),
		Context:  resp.GetContext(),
	}, nil
}

func extractCheck(req *dsa.EvaluationRequest) *dsr.CheckRequest {
	checkReq := &dsr.CheckRequest{}
	if res := req.GetResource(); res != nil {
		checkReq.ObjectType = res.GetType()
		checkReq.ObjectId = res.GetId()
	}

	if act := req.GetAction(); act != nil {
		checkReq.Relation = act.GetName()
	}

	if sub := req.GetSubject(); sub != nil {
		checkReq.SubjectType = sub.GetType()
		checkReq.SubjectId = sub.GetId()
	}

	return checkReq
}

// Evaluations access check.
//
// The Access Evaluations API defines the message exchange pattern between a client (PEP)
// and an authorization service (PDP) for evaluating multiple access evaluations within
// the scope of a single message exchange (also known as "boxcarring" requests).
func (s *Access) Evaluations(ctx context.Context, req *dsa.EvaluationsRequest) (*dsa.EvaluationsResponse, error) {
	defCheck, checks := extractChecks(req)

	checksResp, err := s.reader.Checks(ctx, &dsr.ChecksRequest{Default: defCheck, Checks: checks})
	if err != nil {
		return &dsa.EvaluationsResponse{}, err
	}

	return &dsa.EvaluationsResponse{
		Evaluations: extractDecisions(checksResp),
	}, nil
}

func extractChecks(req *dsa.EvaluationsRequest) (*dsr.CheckRequest, []*dsr.CheckRequest) {
	check := &dsr.CheckRequest{}

	if sub := req.GetSubject(); sub != nil {
		check.SubjectType = sub.GetType()
		check.SubjectId = sub.GetId()
	}

	if act := req.GetAction(); act != nil {
		check.Relation = act.GetName()
	}

	if res := req.GetResource(); res != nil {
		check.ObjectType = res.GetType()
		check.ObjectId = res.GetId()
	}

	checks := make([]*dsr.CheckRequest, len(req.GetEvaluations()))

	for k, v := range req.GetEvaluations() {
		c := extractCheck(v)
		checks[k] = c
	}

	return check, checks
}

func extractDecisions(resp *dsr.ChecksResponse) []*dsa.EvaluationResponse {
	evaluations := make([]*dsa.EvaluationResponse, len(resp.GetChecks()))

	for k, v := range resp.GetChecks() {
		e := &dsa.EvaluationResponse{}
		e.Decision = v.GetCheck()

		if v.GetContext() != nil {
			e.Context = v.GetContext()
		}

		evaluations[k] = e
	}

	return evaluations
}

// SubjectSearch
//
// The Subject Search API defines the message exchange pattern between a client (PEP) and an authorization service (PDP)
// for returning all of the subjects that match the search criteria.
//
// The Subject Search API is based on the Access Evaluation information model, but omits the Subject ID.
func (s *Access) SubjectSearch(ctx context.Context, req *dsa.SubjectSearchRequest) (*dsa.SubjectSearchResponse, error) {
	resp := &dsa.SubjectSearchResponse{
		Results: []*dsa.Subject{},
		Page:    &dsa.PaginationResponse{},
	}

	graphResp, err := s.reader.GetGraph(ctx, extractSubjectSearch(req))
	if err != nil {
		return resp, err
	}

	for _, oid := range graphResp.GetResults() {
		sub := &dsa.Subject{
			Type: oid.GetObjectType(),
			Id:   oid.GetObjectId(),
		}
		resp.Results = append(resp.GetResults(), sub)
	}

	return resp, nil
}

func extractSubjectSearch(req *dsa.SubjectSearchRequest) *dsr.GetGraphRequest {
	resp := &dsr.GetGraphRequest{}
	if res := req.GetResource(); res != nil {
		resp.ObjectType = res.GetType()
		resp.ObjectId = res.GetId()
	}

	if act := req.GetAction(); act != nil {
		resp.Relation = act.GetName()
	}

	if sub := req.GetSubject(); sub != nil {
		resp.SubjectType = sub.GetType()
	}

	resp.SubjectId = ""       // OMITTED
	resp.SubjectRelation = "" // OMITTED

	return resp
}

// ResourceSearch
//
// The Resource Search API defines the message exchange pattern between a client (PEP) and an authorization service (PDP)
// for returning all of the resources that match the search criteria.
//
// The Resource Search API is based on the Access Evaluation information model, but omits the Resource ID.
func (s *Access) ResourceSearch(ctx context.Context, req *dsa.ResourceSearchRequest) (*dsa.ResourceSearchResponse, error) {
	resp := &dsa.ResourceSearchResponse{
		Results: []*dsa.Resource{},
		Page:    &dsa.PaginationResponse{},
	}

	graphResp, err := s.reader.GetGraph(ctx, extractResourceSearch(req))
	if err != nil {
		return resp, err
	}

	for _, oid := range graphResp.GetResults() {
		res := &dsa.Resource{
			Type: oid.GetObjectType(),
			Id:   oid.GetObjectId(),
		}

		resp.Results = append(resp.GetResults(), res)
	}

	return resp, nil
}

func extractResourceSearch(req *dsa.ResourceSearchRequest) *dsr.GetGraphRequest {
	resp := &dsr.GetGraphRequest{}
	if res := req.GetResource(); res != nil {
		resp.ObjectType = res.GetType()
	}

	if act := req.GetAction(); act != nil {
		resp.Relation = act.GetName()
	}

	if sub := req.GetSubject(); sub != nil {
		resp.SubjectType = sub.GetType()
		resp.SubjectId = sub.GetId()
	}

	resp.ObjectId = ""        // OMITTED
	resp.SubjectRelation = "" // OMITTED

	return resp
}

// ActionSearch
//
// The Action Search API defines the message exchange pattern between a client (PEP) and an authorization service (PDP)
// for returning all of the actions that match the search criteria.
//
// The Action Search API is based on the Access Evaluation information model.
func (s *Access) ActionSearch(ctx context.Context, req *dsa.ActionSearchRequest) (*dsa.ActionSearchResponse, error) {
	resp := &dsa.ActionSearchResponse{
		Results: []*dsa.Action{},
		Page:    &dsa.PaginationResponse{},
	}

	graphReq := extractActionSearch(req)
	inclRelations := false

	assignable := []cache.RelationName{}

	if inclRelations {
		assignableRelations, err := s.reader.store.MC().AssignableRelations(
			cache.ObjectName(graphReq.GetObjectType()),
			cache.ObjectName(graphReq.GetSubjectType()),
		)
		if err != nil {
			return resp, err
		}

		assignable = append(assignable, assignableRelations...)
	}

	availablePermissions, err := s.reader.store.MC().AvailablePermissions(
		cache.ObjectName(graphReq.GetObjectType()),
		cache.ObjectName(graphReq.GetSubjectType()),
	)
	if err != nil {
		return resp, err
	}

	assignable = append(assignable, availablePermissions...)

	checks := []*dsr.CheckRequest{}
	for _, rel := range assignable {
		checks = append(checks, &dsr.CheckRequest{Relation: rel.String()})
	}

	checksResp, err := s.reader.Checks(ctx, &dsr.ChecksRequest{
		Default: &dsr.CheckRequest{
			ObjectType:  graphReq.GetObjectType(),
			ObjectId:    graphReq.GetObjectId(),
			SubjectType: graphReq.GetSubjectType(),
			SubjectId:   graphReq.GetSubjectId(),
		},
		Checks: checks,
	})
	if err != nil {
		return resp, err
	}

	for i, chk := range checksResp.GetChecks() {
		if chk.GetCheck() {
			resp.Results = append(resp.GetResults(), &dsa.Action{Name: assignable[i].String()})
		}
	}

	return resp, nil
}

func extractActionSearch(req *dsa.ActionSearchRequest) *dsr.GetGraphRequest {
	resp := &dsr.GetGraphRequest{}
	if res := req.GetResource(); res != nil {
		resp.ObjectType = res.GetType()
		resp.ObjectId = res.GetId()
	}

	if sub := req.GetSubject(); sub != nil {
		resp.SubjectType = sub.GetType()
		resp.SubjectId = sub.GetId()
	}

	resp.Relation = ""        // OMITTED
	resp.SubjectRelation = "" // OMITTED

	return resp
}
