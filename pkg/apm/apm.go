package apm

import (
	"bytes"
	apm "go.elastic.co/apm/model"
	"go.elastic.co/fastjson"
	"log"
	"net/http"
	"net/url"
)

const (
	apmUrl = "http://172.17.0.5:8200/intake/v2/events"
)

func SendToAPM(transaction *apm.Transaction) error {
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

	_, err := http.Post(apmUrl, "application/x-ndjson", buf)
	return err
}

func UrlToAPM(requestUrl url.URL) apm.URL {
	return apm.URL{
		Full:     requestUrl.String(),
		Protocol: requestUrl.Scheme,
		Hostname: requestUrl.Hostname(),
		Port:     requestUrl.Port(),
		Path:     requestUrl.Path,
		Search:   requestUrl.RawQuery,
		Hash:     requestUrl.Fragment,
	}
}
