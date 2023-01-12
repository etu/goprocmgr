package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Serve struct {
	config *Config
}

func (serve *Serve) Run(config *Config) {
	serve.config = config

	router := serve.newRouter()

	log.Printf("Listening on http://%s:%d\n", serve.config.Settings.ListenAddress, serve.config.Settings.ListenPort)

	// Listen to configured address and port.
	log.Fatal(http.ListenAndServe(
		fmt.Sprintf("%s:%d", serve.config.Settings.ListenAddress, serve.config.Settings.ListenPort),
		router,
	))
}

func (serve *Serve) newRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	// Web server endpoints
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})
	router.HandleFunc("/web/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/style.css")
	})
	router.HandleFunc("/web/script.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/script.js")
	})

	// This endpoint is served at GET /api/config and it returns the
	// currently loaded config.
	router.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(serve.config)
	}).Methods("GET")

	router.HandleFunc("/api/config/server", func(w http.ResponseWriter, r *http.Request) {
		var server ServerConfig
		resp := make(map[string]string)

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&server)

		if err != nil {
			w.WriteHeader(400)
			resp["message"] = fmt.Sprintf("%s", err)
		} else {
			resp["message"] = "OK"

			// Store the sent server config to the config.
			serve.config.Servers[server.Name] = server

			// Save the config to disk.
			serve.config.Save()
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}).Methods("POST")

	return router
}
