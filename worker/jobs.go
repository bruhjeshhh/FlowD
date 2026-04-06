package worker

import "log/slog"

func handleemails(log *slog.Logger, payload []byte) bool {
	log.Info("handling email job", "payload_len", len(payload))
	return true
}

func handlesms(log *slog.Logger, payload []byte) bool {
	log.Info("handling sms job", "payload_len", len(payload))
	return true
}

func handlepushnotifications(log *slog.Logger, payload []byte) bool {
	log.Info("handling push_notification job", "payload_len", len(payload))
	return true
}
