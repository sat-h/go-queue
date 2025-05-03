package main

import (
	"log"
	"net/http"

	"github.com/sat-h/go-queue/internal/api"
	"github.com/sat-h/go-queue/internal/job"
)

func main() {
	queue := job.NewQueue("localhost:6379")
	handler := &api.Handler{Queue: queue}
	router := api.NewRouter(handler)

	log.Println("API running on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
