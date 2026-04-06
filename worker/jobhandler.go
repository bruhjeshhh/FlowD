package worker

import "log/slog"

func handlejobs(log *slog.Logger, jobtype string, payload []byte) bool {
	if log == nil {
		log = slog.Default()
	}
	switch jobtype {
	case "email":
		return handleemails(log, payload)
	case "sms":
		return handlesms(log, payload)
	case "push_notification":
		return handlepushnotifications(log, payload)
	default:
		return false
	}
}
