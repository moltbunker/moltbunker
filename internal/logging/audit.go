package logging

// AuditEvent represents a sensitive operation that should be logged for compliance
type AuditEvent struct {
	Operation string // e.g., "key_generated", "stake_created", "container_deployed"
	Actor     string // Who performed the action (node ID, wallet address, API key prefix)
	Target    string // What was affected (provider address, container ID, etc.)
	Result    string // "success" or "failure"
	Details   string // Additional context
}

// Audit logs a sensitive operation with structured fields.
// Audit events are logged at Info level with a special "audit" attribute
// to distinguish them from regular application logs.
func Audit(event AuditEvent) {
	Logger().Info("audit",
		"audit", true,
		"operation", event.Operation,
		"actor", event.Actor,
		"target", event.Target,
		"result", event.Result,
		"details", event.Details,
	)
}
