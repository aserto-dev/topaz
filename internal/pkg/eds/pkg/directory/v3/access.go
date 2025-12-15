package v3

import (
	"context"

	"github.com/aserto-dev/azm/cache"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	acc1 "github.com/authzen/access.go/api/access/v1"
	"github.com/rs/zerolog"
)

type Access struct {
	acc1.UnimplementedAccessServer

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
func (s *Access) Evaluation(ctx context.Context, req *acc1.EvaluationRequest) (*acc1.EvaluationResponse, error) {
	resp, err := s.reader.Check(ctx, extractCheck(req))
	if err != nil {
		return &acc1.EvaluationResponse{}, err
	}

	return &acc1.EvaluationResponse{
		Decision: resp.GetCheck(),
		Context:  resp.GetContext(),
	}, nil
}

func extractCheck(req *acc1.EvaluationRequest) *dsr3.CheckRequest {
	checkReq := &dsr3.CheckRequest{}
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
func (s *Access) Evaluations(ctx context.Context, req *acc1.EvaluationsRequest) (*acc1.EvaluationsResponse, error) {
	defCheck, checks := extractChecks(req)

	checksResp, err := s.reader.Checks(ctx, &dsr3.ChecksRequest{Default: defCheck, Checks: checks})
	if err != nil {
		return &acc1.EvaluationsResponse{}, err
	}

	return &acc1.EvaluationsResponse{
		Evaluations: extractDecisions(checksResp),
	}, nil
}

func extractChecks(req *acc1.EvaluationsRequest) (*dsr3.CheckRequest, []*dsr3.CheckRequest) {
	check := &dsr3.CheckRequest{}

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

	checks := make([]*dsr3.CheckRequest, len(req.GetEvaluations()))

	for k, v := range req.GetEvaluations() {
		c := extractCheck(v)
		checks[k] = c
	}

	return check, checks
}

func extractDecisions(resp *dsr3.ChecksResponse) []*acc1.EvaluationResponse {
	evaluations := make([]*acc1.EvaluationResponse, len(resp.GetChecks()))

	for k, v := range resp.GetChecks() {
		e := &acc1.EvaluationResponse{}
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
func (s *Access) SubjectSearch(ctx context.Context, req *acc1.SubjectSearchRequest) (*acc1.SubjectSearchResponse, error) {
	resp := &acc1.SubjectSearchResponse{
		Results: []*acc1.Subject{},
		Page:    &acc1.PaginationResponse{},
	}

	graphResp, err := s.reader.GetGraph(ctx, extractSubjectSearch(req))
	if err != nil {
		return resp, err
	}

	for _, oid := range graphResp.GetResults() {
		sub := &acc1.Subject{
			Type: oid.GetObjectType(),
			Id:   oid.GetObjectId(),
		}
		resp.Results = append(resp.GetResults(), sub)
	}

	return resp, nil
}

func extractSubjectSearch(req *acc1.SubjectSearchRequest) *dsr3.GetGraphRequest {
	resp := &dsr3.GetGraphRequest{}
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
func (s *Access) ResourceSearch(ctx context.Context, req *acc1.ResourceSearchRequest) (*acc1.ResourceSearchResponse, error) {
	resp := &acc1.ResourceSearchResponse{
		Results: []*acc1.Resource{},
		Page:    &acc1.PaginationResponse{},
	}

	graphResp, err := s.reader.GetGraph(ctx, extractResourceSearch(req))
	if err != nil {
		return resp, err
	}

	for _, oid := range graphResp.GetResults() {
		res := &acc1.Resource{
			Type: oid.GetObjectType(),
			Id:   oid.GetObjectId(),
		}

		resp.Results = append(resp.GetResults(), res)
	}

	return resp, nil
}

func extractResourceSearch(req *acc1.ResourceSearchRequest) *dsr3.GetGraphRequest {
	resp := &dsr3.GetGraphRequest{}
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
func (s *Access) ActionSearch(ctx context.Context, req *acc1.ActionSearchRequest) (*acc1.ActionSearchResponse, error) {
	resp := &acc1.ActionSearchResponse{
		Results: []*acc1.Action{},
		Page:    &acc1.PaginationResponse{},
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

	checks := []*dsr3.CheckRequest{}
	for _, rel := range assignable {
		checks = append(checks, &dsr3.CheckRequest{Relation: rel.String()})
	}

	checksResp, err := s.reader.Checks(ctx, &dsr3.ChecksRequest{
		Default: &dsr3.CheckRequest{
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
			resp.Results = append(resp.GetResults(), &acc1.Action{Name: assignable[i].String()})
		}
	}

	return resp, nil
}

func extractActionSearch(req *acc1.ActionSearchRequest) *dsr3.GetGraphRequest {
	resp := &dsr3.GetGraphRequest{}
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
