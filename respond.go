package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorstc struct {
		Error string `json:"error"`
	}
	errormsg := errorstc{
		Error: msg,
	}

	resp, eror := json.Marshal(errormsg)
	if eror != nil {
		log.Printf("cant unmarshal the errormsg check respondwitherror func")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(resp)

}

func respondWithJson(w http.ResponseWriter, n int, payload any) {

	resp, eror := json.Marshal(payload)
	if eror != nil {
		log.Printf("cant marshal the payload check respondwithJson func")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(n)
	w.Write(resp)

}
