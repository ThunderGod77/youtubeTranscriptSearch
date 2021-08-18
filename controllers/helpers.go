package controllers

import (
	"encoding/json"
	"log"
	"net/http"
)


type webError struct {
	Msg        string `json:"msg"`
	Err        bool   `json:"err"`
	statusCode int
}

func (we webError) ReturnError(w http.ResponseWriter) {
	log.Println(we.Msg)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(we.statusCode)
	jE, _ := json.Marshal(we)
	_, err := w.Write(jE)
	if err != nil {
		log.Println(err)
	}

}

// to handle and return errors to client
func newWebError(w http.ResponseWriter, err error, statusCode int) {
	we := webError{
		Msg:        err.Error(),
		Err:        true,
		statusCode: statusCode,
	}
	we.ReturnError(w)
}

func sendResp(w http.ResponseWriter, statusCode int, respJs []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(respJs)
}
