package graph

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
)

type checkStatus int

const (
	checkStatusNew checkStatus = iota
	checkStatusPending
	checkStatusTrue
	checkStatusFalse
)

const (
	statusUnknown  = "new"
	statusPending  = "?"
	statusTrue     = "true"
	statusFalse    = "false"
	statusComplete = "done"
)

func (s checkStatus) String() string {
	switch s {
	case checkStatusNew:
		return statusUnknown
	case checkStatusPending:
		return statusPending
	case checkStatusTrue:
		return statusTrue
	case checkStatusFalse:
		return statusFalse
	default:
		return fmt.Sprintf("invalid: %d", s)
	}
}

type checkCall struct {
	*relation

	status checkStatus
}

// checkMemo tracks pending checks to detect cycles, and caches the results of completed checks.
type checkMemo struct {
	memo    map[relation]checkStatus
	history []*checkCall
	cycles  []*relation
}

func newCheckMemo(trace bool) *checkMemo {
	return &checkMemo{
		memo:    map[relation]checkStatus{},
		history: lo.Ternary(trace, []*checkCall{}, nil),
	}
}

// MarkVisited returns the status of a check. If the check has not been visited, it is marked as pending.
func (m *checkMemo) MarkVisited(params *relation) checkStatus {
	prior := m.memo[*params]
	current := prior

	switch prior {
	case checkStatusNew:
		current = checkStatusPending
		m.memo[*params] = current
	case checkStatusPending:
		m.cycles = append(m.cycles, params)
	case checkStatusTrue, checkStatusFalse:
		break
	}

	m.trace(params, current)

	return prior
}

// MarkComplete records the result of a check.
func (m *checkMemo) MarkComplete(params *relation, status checkStatus) {
	m.memo[*params] = status

	m.trace(params, status)
}

func (m *checkMemo) Trace() []string {
	if m.history == nil {
		return []string{}
	}

	callstack := []string{}

	return lo.Map(m.history, func(c *checkCall, _ int) string {
		call := c.String()
		result := c.status.String()

		if len(callstack) > 0 && callstack[len(callstack)-1] == call && c.status != checkStatusPending {
			callstack = callstack[:len(callstack)-1]
		}

		s := fmt.Sprintf("%s%s = %s", strings.Repeat("  ", len(callstack)), call, result)

		if c.status == checkStatusPending {
			callstack = append(callstack, call)
		}

		return s
	})
}

func (m *checkMemo) trace(params *relation, status checkStatus) {
	if m.history != nil {
		m.history = append(m.history, &checkCall{params, status})
	}
}
