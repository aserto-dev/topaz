package builtins

const (
	ds  string = "ds"
	az  string = "az"
	dot string = "."
)

const (
	Check           string = "check"
	Checks          string = "checks"
	Graph           string = "graph"
	Object          string = "object"
	Relation        string = "relation"
	Relations       string = "relations"
	Resolve         string = "resolve"
	Identity        string = "identity"
	User            string = "user"
	CheckRelation   string = "check_relation"
	CheckPermission string = "check_permission"
)

// Topaz Directory built-ins.
const (
	DSCheck           string = ds + dot + Check           // ds.check
	DSChecks          string = ds + dot + Checks          // ds.checks
	DSGraph           string = ds + dot + Graph           // ds.graph
	DSObject          string = ds + dot + Object          // ds.object
	DSRelation        string = ds + dot + Relation        // ds.relation
	DSRelations       string = ds + dot + Relations       // ds.relations
	DSResolve         string = ds + dot + Resolve         // ds.resolve
	DSIdentity        string = ds + dot + Identity        // ds.identity (OBSOLETE, use `ds.resolve`)
	DSUser            string = ds + dot + User            // ds.user (OBSOLETE, use `ds.object``)
	DSCheckRelation   string = ds + dot + CheckRelation   // ds.check_relation (OBSOLETE, use `ds.check`)
	DSCheckPermission string = ds + dot + CheckPermission // ds.check_permission (OBSOLETE, use `ds.check`)
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
