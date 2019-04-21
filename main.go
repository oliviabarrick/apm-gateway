package main

import (
	"log"
	"net/http"
	"github.com/justinbarrick/apm-gateway/pkg/importers/zipkin"
)

func main() {
	http.HandleFunc("/api/v2/spans", zipkin.Handler)
	log.Fatal(http.ListenAndServe(":9411", nil))
}
