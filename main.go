package main

import (
	"github.com/justinbarrick/apm-gateway/pkg/importers/zipkin"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/api/v2/spans", zipkin.Handler)
	log.Fatal(http.ListenAndServe(":9411", nil))
}
