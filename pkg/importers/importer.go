package importers

import (
	"github.com/justinbarrick/apm-gateway/pkg/exporters"
	"net/http"
)

type Importer interface {
	SetExporter(e exporter.Exporter)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

func WithExporter(i Importer, e exporter.Exporter) Importer {
	i.SetExporter(e)
	return i
}
