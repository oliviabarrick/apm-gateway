package main

import (
	"github.com/justinbarrick/apm-gateway/pkg/importers/jaeger"
	"github.com/justinbarrick/apm-gateway/pkg/importers/zipkin"
	"log"
	"net/http"
)

func main() {
	go func() {
		log.Fatal(http.ListenAndServe(":9411", http.HandlerFunc(zipkin.Handler)))
	}()

	go func() {
		log.Fatal(http.ListenAndServe(":14268", http.HandlerFunc(jaeger.Handler)))
	}()

	select {}
}
