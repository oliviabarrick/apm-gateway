package main

import (
	"github.com/justinbarrick/apm-gateway/pkg/exporters/apm"
	"github.com/justinbarrick/apm-gateway/pkg/importers/jaeger"
	"github.com/justinbarrick/apm-gateway/pkg/importers/zipkin"
	"log"
	"net/http"
)

func main() {
	exporter := apm.Exporter{}

	go func() {
		log.Fatal(http.ListenAndServe(":9411", importer.WithExporter(&zipkin.Importer{}, exporter)))
	}()

	go func() {
		log.Fatal(http.ListenAndServe(":14268", importer.WithExporter(&jaeger.Importer{}, exporter)))
	}()

	select {}
}
