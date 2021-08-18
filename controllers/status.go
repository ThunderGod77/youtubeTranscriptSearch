package controllers

import (
	"encoding/json"
	"errors"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"net/http"
	"rssFeedSearch/global"
)

type statusResp struct {
	Msg    string `json:"msg"`
	Status string `json:"status"`
}

func getProcessStatus(id string) (statusResp, error) {
	conn := global.ConnPool.Get()
	defer conn.Close()

	info, err := redis.StringMap(conn.Do("HGETALL", id))
	if err != nil {
		return statusResp{}, err
	}
	status := info["status"]
	if status != "" {
		return statusResp{Msg: info["msg"], Status: info["status"]}, err
	} else {
		return statusResp{}, errors.New("process id does not exist")
	}

}

func CheckStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	st, err := getProcessStatus(id)
	if err != nil {
		newWebError(w, err, http.StatusBadRequest)
	}

	resp, err := json.Marshal(st)
	if err != nil {
		newWebError(w, err, http.StatusInternalServerError)
		return
	}
	sendResp(w, http.StatusOK, resp)
	return

}
