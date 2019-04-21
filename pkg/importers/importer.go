package importer

import (
	"github.com/justinbarrick/apm-gateway/pkg/exporters"
	"log"
	"net/http"
)

type Importer interface {
	SetExporter(e exporter.Exporter)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

func Serve(port string, i Importer, e exporter.Exporter) {
	i.SetExporter(e)

	go func() {
		log.Fatal(http.ListenAndServe(port, i))
	}()
}
