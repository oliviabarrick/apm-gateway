package main

import (
	"github.com/justinbarrick/apm-gateway/pkg/exporters/apm"
	"github.com/justinbarrick/apm-gateway/pkg/importers"
	"github.com/justinbarrick/apm-gateway/pkg/importers/jaeger"
	"github.com/justinbarrick/apm-gateway/pkg/importers/zipkin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
)

func main() {
	exporter := &apm.Exporter{Url: os.Getenv("APM_ENDPOINT")}
	importer.Serve(":9411", &zipkin.Importer{}, exporter)
	importer.Serve(":14268", &jaeger.Importer{}, exporter)
	log.Fatal(http.ListenAndServe(":8080", promhttp.Handler()))
}
