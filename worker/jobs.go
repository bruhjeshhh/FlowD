package worker

import "log"

func handleemails(payload []byte) bool {
	log.Printf("Handling email job" + string(payload))
	return true
}

func handlesms(payload []byte) bool {
	log.Printf("Handling SMS job" + string(payload))
	return true
}

func handlepushnotifications(payload []byte) bool {
	log.Printf("Handling push notification job" + string(payload))
	return true
}
