package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	settings := loadSettingsFromFile("config.json")

	handler := HookHandler{conf: settings}
	s := &http.Server{
		Addr:           settings.Server.Interface + ":" + settings.Server.Port,
		Handler:        &handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	fmt.Printf("Serving on %s:%s\n", settings.Server.Interface, settings.Server.Port)
	err := s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
