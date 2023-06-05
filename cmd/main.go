package main

import (
	"command-on-demand/internal/logger"
	s "command-on-demand/internal/server"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const (
	pruneInterval = 10 * time.Second
	writeTimeout  = 15 * time.Second
	readTimeout   = 15 * time.Second
	idleTimeout   = 60 * time.Second
)

func main() {
	logger.Setup()
	r := mux.NewRouter()
	srv := s.NewServer()
	go srv.CodeStore.Prune(pruneInterval)
	r.HandleFunc("/api/v1/code/{udid}", srv.CodeHandler).Methods("GET")
	r.HandleFunc("/api/v1/erase/{udid}", srv.EraseHandler).Methods("POST")
	r.HandleFunc("/api/v1/swupd/{udid}", srv.SoftwareUpdateHandler).Methods("POST")

	// skip middleware and obfuscate 404 with 403 for unknown paths
	nfh := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		return
	})
	r.NotFoundHandler = nfh

	r.Use(s.MiddlewareSetRequestId)
	r.Use(s.MiddlewareLogging)
	r.Use(srv.MiddlewareBearerAuth)

	addr := fmt.Sprintf("%s:%s", srv.ListenInterface(), srv.ListenPort())
	hs := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
		IdleTimeout:  idleTimeout,
	}

	logger.Info("service running and listening on ", addr)
	logger.Fatal(hs.ListenAndServe())
}
