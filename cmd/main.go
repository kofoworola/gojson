package main

import (
	"log"
	"net/http"
)

func main() {
	handler, err := NewHandler()
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", handler)
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}

}
