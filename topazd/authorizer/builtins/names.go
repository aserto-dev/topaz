package builtins

const (
	ds  string = "ds"
	az  string = "az"
	dot string = "."
)

const (
	Check     string = "check"
	Checks    string = "checks"
	Graph     string = "graph"
	Object    string = "object"
	Relation  string = "relation"
	Relations string = "relations"
	Identity  string = "identity"
	User      string = "user"
)

// Topaz Directory built-ins.
const (
	DSCheck     string = ds + dot + Check     // ds.check
	DSChecks    string = ds + dot + Checks    // ds.checks
	DSGraph     string = ds + dot + Graph     // ds.graph
	DSObject    string = ds + dot + Object    // ds.object
	DSRelation  string = ds + dot + Relation  // ds.relation
	DSRelations string = ds + dot + Relations // ds.relations
	DSIdentity  string = ds + dot + Identity  // ds.identity (OBSOLETE)
	DSUser      string = ds + dot + User      // ds.user (OBSOLETE)
)

const (
	Evaluation     string = "evaluation"
	Evaluations    string = "evaluations"
	SubjectSearch  string = "subject_search"
	ResourceSearch string = "resource_search"
	ActionSearch   string = "action_search"
)

// OpenID AuthZEN built-ins, see https://openid.github.io/authzen/
const (
	AZEvaluation     string = az + dot + Evaluation     // az.evaluation
	AZEvaluations    string = az + dot + Evaluations    // az.evaluations
	AZSubjectSearch  string = az + dot + SubjectSearch  // az.subject_search
	AZResourceSearch string = az + dot + ResourceSearch // az.resource_search
	AZActionSearch   string = az + dot + ActionSearch   // az.action_search
)
