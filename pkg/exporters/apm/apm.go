package apm

import (
	"bytes"
	apm "go.elastic.co/apm/model"
	"go.elastic.co/fastjson"
	"log"
	"net/http"
)

const (
	apmUrl = "http://172.17.0.5:8200/intake/v2/events"
)

type Exporter struct {
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

	_, err := e.client.Post(apmUrl, "application/x-ndjson", buf)
	return err
}
