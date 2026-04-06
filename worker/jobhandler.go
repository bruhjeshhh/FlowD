package worker

func handlejobs(jobtype string, payload []byte) bool {
	switch jobtype {
	case "email":
		return handleemails(payload)
	case "sms":
		return handlesms(payload)
	case "push_notification":
		return handlepushnotifications(payload)
	default:
		return false
	}
}
