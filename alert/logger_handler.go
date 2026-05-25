package alert

import "github.com/user/procwatch/logger"

// LoggerHandler returns an alert Handler that writes events to the
// provided structured logger using its Alert method. The severity level
// of the log entry mirrors the event's Severity field, so critical alerts
// are distinguishable from warnings in the log output.
func LoggerHandler(log *logger.Logger) Handler {
	return func(e Event) {
		fields := map[string]interface{}{
			"process":   e.Process,
			"pid":       e.PID,
			"metric":    e.Metric,
			"value":     e.Value,
			"threshold": e.Threshold,
			"severity":  string(e.Severity),
		}
		switch e.Severity {
		case SeverityCritical:
			log.Error(e.Message, fields)
		case SeverityWarning:
			log.Warn(e.Message, fields)
		default:
			log.Alert(e.Message, fields)
		}
	}
}
