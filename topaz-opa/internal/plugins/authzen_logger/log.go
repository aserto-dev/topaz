package authzen_logger

import (
	"context"
	"encoding/json"

	"github.com/open-policy-agent/opa/v1/plugins/logs"
)

// AuthZenEvaluation represents the AuthZEN standard 1.0 schema structure.
type AuthZenEvaluation struct {
	DecisionID string         `json:"decision_id"`
	Timestamp  string         `json:"timestamp"`
	Subject    any            `json:"subject"`  // P - Principal / Subject
	Action     any            `json:"action"`   // A - Action
	Resource   any            `json:"resource"` // R - Resource
	Context    any            `json:"context"`  // C - Context
	Decision   bool           `json:"decision"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// Log intercepts OPA events and processes them.
func (l *Plugin) Log(ctx context.Context, event logs.EventV1) error {
	// Parse OPA's raw input payload
	var inputMap map[string]any

	inputBytes, _ := json.Marshal(event.Input) //nolint:errchkjson
	_ = json.Unmarshal(inputBytes, &inputMap)

	// Extract the actual boolean evaluation result safely
	var decisionResult bool

	if event.Result != nil { //nolint:nestif
		var res bool

		resBytes, _ := json.Marshal(*event.Result) //nolint:errchkjson
		if err := json.Unmarshal(resBytes, &res); err == nil {
			decisionResult = res
		} else {
			// If it returns a complex map structure, check standard rule fields like 'allow'
			var resMap map[string]any

			_ = json.Unmarshal(resBytes, &resMap)
			if allow, exists := resMap["allow"]; exists {
				if decision, ok := allow.(bool); ok {
					decisionResult = decision
				}
			}
		}
	}

	const ns2msDiv int64 = 1000000 // nano seconds to milliseconds divider.

	// Extract execution latency safely from the map
	var ndMs int64

	if event.Metrics != nil {
		// OPA natively populates specific timer keys as float64 when unmarshaled or stored as interface values
		if evalNs, ok := event.Metrics["timer_rego_query_eval_ns"].(float64); ok {
			ndMs = int64(evalNs) / ns2msDiv // Convert nanoseconds to milliseconds
		} else if evalNsInt, ok := event.Metrics["timer_rego_query_eval_ns"].(int64); ok {
			ndMs = evalNsInt / ns2msDiv
		}
	}

	// Format into AuthZEN PARC layout
	parcLog := AuthZenEvaluation{
		DecisionID: event.DecisionID,
		Timestamp:  event.Timestamp.String(),
		Subject:    inputMap["subject"],
		Action:     inputMap["action"],
		Resource:   inputMap["resource"],
		Context:    inputMap["context"],
		Decision:   decisionResult,
		Metadata: map[string]any{
			"path":  event.Path,
			"nd_ms": ndMs,
		},
	}

	// Serialize object to JSON line
	logLine, err := json.Marshal(parcLog)
	if err != nil {
		return err
	}

	// Write directly via lumberjack thread-safe file handling
	_, err = l.fileLogger.Write(append(logLine, '\n'))

	return err
}
