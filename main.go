package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"rssFeedSearch/controllers"
	"rssFeedSearch/global"
)

func main() {

	//initializing redis connection pool
	global.RedisInit()
	global.ElasticInit()
	global.DeleteAllRedis()



	r := mux.NewRouter()

	//subrouter to handle search capabilities
	s := r.PathPrefix("/search").Subrouter()
	s.StrictSlash(true)

	//subtouter to handle video insertion routes
	a := r.PathPrefix("/add").Subrouter()
	a.StrictSlash(true)

	//test route
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		t, err := json.Marshal(map[string]string{"ping": "pong"})

		_, err = w.Write(t)
		if err != nil {
			log.Println(err)
		}
	}).Methods("GET")

	//route to add videos
	a.HandleFunc("", controllers.AddVideos)
	//route to check status
	r.HandleFunc("/status/{id}", controllers.CheckStatus)
	//route to handle search
	s.HandleFunc("", controllers.SearchByContent)

	log.Println("Starting server on port 8080!")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}

}
