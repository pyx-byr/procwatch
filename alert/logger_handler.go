package alert

import "github.com/user/procwatch/logger"

// LoggerHandler returns an alert Handler that writes events to the
// provided structured logger using its Alert method.
func LoggerHandler(log *logger.Logger) Handler {
	return func(e Event) {
		log.Alert(
			e.Message,
			map[string]interface{}{
				"process":   e.Process,
				"pid":       e.PID,
				"metric":    e.Metric,
				"value":     e.Value,
				"threshold": e.Threshold,
				"severity":  string(e.Severity),
			},
		)
	}
}
