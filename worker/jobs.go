package worker

import "log/slog"

func handlesms(log *slog.Logger, payload []byte) (bool, error) {
	log.Info("handling sms job", "payload_len", len(payload))
	return true, nil
}

func handlepushnotifications(log *slog.Logger, payload []byte) (bool, error) {
	log.Info("handling push_notification job", "payload_len", len(payload))
	return true, nil
}
