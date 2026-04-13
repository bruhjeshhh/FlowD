package worker

import "log/slog"

type JobResult struct {
	Success bool
	Error   error
}

func handlejobs(log *slog.Logger, jobtype string, payload []byte) JobResult {
	if log == nil {
		log = slog.Default()
	}
	switch jobtype {
	case "email":
		ok, err := handleemails(log, payload)
		return JobResult{Success: ok, Error: err}
	case "sms":
		ok, err := handlesms(log, payload)
		return JobResult{Success: ok, Error: err}
	case "push_notification":
		ok, err := handlepushnotifications(log, payload)
		return JobResult{Success: ok, Error: err}
	default:
		return JobResult{Success: false, Error: nil}
	}
}
