package apm

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
	"bytes"
	apm "go.elastic.co/apm/model"
	"go.elastic.co/fastjson"
	"log"
	"net/http"
)

var (
	exportSeconds = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "apm_export_seconds",
			Help:       "Seconds exporting traces to APM.",
		},
		[]string{"service", "status"},
	)

	exportErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:       "apm_export_errors",
			Help:       "Errors observed exporting traces to APM.",
		},
		[]string{"service"},
	)
)

type Exporter struct {
	Url string
	client *http.Client
}

func (e *Exporter) SendToAPM(transaction *apm.Transaction) error {
	if e.client == nil {
		e.client = &http.Client{}
	}

	var transactionEncoded fastjson.Writer
	fastjson.Marshal(&transactionEncoded, transaction)

	var metadata fastjson.Writer
	fastjson.Marshal(&metadata, &apm.Service{
		Name: "apm-gateway",
		Agent: &apm.Agent{
			Name:    "apm-gateway",
			Version: "0.0.1",
		},
	})

	buf := &bytes.Buffer{}
	buf.Write([]byte("{\"metadata\":{\"service\":"))
	buf.Write(metadata.Bytes())
	buf.Write([]byte("}}\n{\"transaction\":"))
	buf.Write(transactionEncoded.Bytes())
	buf.Write([]byte("}\n"))

	log.Println(string(buf.Bytes()))

	startTime := time.Now()

	resp, err := e.client.Post(e.Url, "application/x-ndjson", buf)

	service := ""
	if transaction.Context.Service != nil {
		service = transaction.Context.Service.Name
	}
	exportSeconds.WithLabelValues(service, resp.Status).Observe(float64(time.Now().Sub(startTime).Nanoseconds()))

	if err != nil {
		exportErrors.WithLabelValues(service).Inc()
		return err
	}

	if err := resp.Body.Close(); err != nil {
		return err
	}

	return err
}

func init() {
	prometheus.MustRegister(exportSeconds)
	prometheus.MustRegister(exportErrors)
}
