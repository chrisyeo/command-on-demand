package main

import (
	"command-on-demand/internal/logger"
	s "command-on-demand/internal/server"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

func main() {
	logger.Setup()
	r := mux.NewRouter()
	svc := s.NewService()
	go svc.CodeStore.Prune(10 * time.Second)
	r.HandleFunc("/api/v1/code/{udid}", svc.CodeHandler).Methods("GET")
	r.HandleFunc("/api/v1/erase/{udid}", svc.EraseHandler).Methods("POST")
	r.HandleFunc("/api/v1/swupd/{udid}", svc.SoftwareUpdateHandler).Methods("POST")

	// skip middleware and obfuscate 404 with 403 for unknown paths
	nfh := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		return
	})
	r.NotFoundHandler = nfh

	r.Use(s.MiddlewareSetRequestId)
	r.Use(s.MiddlewareLogging)
	r.Use(svc.MiddlewareBearerAuth)

	addr := fmt.Sprintf("%s:%s", svc.ListenInterface(), svc.ListenPort())
	hs := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("service running and listening on ", addr)
	logger.Fatal(hs.ListenAndServe())

}
