package main

import (
	"log"
	"net/http"
	"os"

	httpapi "github.com/AYaSmyslov/faqapi/internal/http"
	"github.com/AYaSmyslov/faqapi/internal/service"
	"github.com/AYaSmyslov/faqapi/internal/storage"
)

func main() {
	db := storage.NewPostgesDB()
	svc := service.NewFAQService(db)
	srv := httpapi.NewServer(svc)

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, httpapi.Logging(srv)); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
