package main

import (
	"bytes"
	"log"
	"net/url"
	"net/http"
	"time"
	"go.elastic.co/fastjson"
	apm "go.elastic.co/apm/model"
	//apmtransport "go.elastic.co/apm/transport"
)

func main() {
	url, err := url.Parse("https://google.com/")
	if err != nil {
		log.Fatal(err)
	}

	transaction := &apm.Transaction{
		ID: apm.SpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 11}),
		TraceID: apm.TraceID([16]byte{
			1, 2, 3, 4, 5, 6, 7, 11, 9, 10, 11, 12, 13, 14, 15, 19,
		}),
		ParentID: apm.SpanID([8]byte{0,0,0,0,0,0,0,0}),
		Name: "myspan",
		Type: "request",
		Timestamp: apm.Time(time.Now()),
		Duration: 1000,
		Result: "200",
		Context: &apm.Context{
			Request: &apm.Request{
				URL: apm.URL{
					Full: url.String(),
					Protocol: url.Scheme,
					Hostname: url.Hostname(),
					Port: url.Port(),
					Path: url.Path,
					Search: url.RawQuery,
					Hash: url.Fragment,
				},
				Method: "GET",
			},
			Response: &apm.Response{
				StatusCode: 200,
			},
			Tags: apm.StringMap{
				apm.StringMapItem{
					Key: "hello",
					Value: "world",
				},
			},
		},
		SpanCount: apm.SpanCount{
			Dropped: 0,
			Started: 0,
		},
	}

	var w fastjson.Writer
	fastjson.Marshal(&w, transaction)

	buf := &bytes.Buffer{}
	buf.Write([]byte("{\"metadata\":{\"service\":{\"name\":\"zipkin2apm\",\"agent\":{\"name\":\"zipkin2apm\",\"version\":\"0.0.1\"}}}}\n{\"transaction\":"))
	buf.Write(w.Bytes())
	buf.Write([]byte("}\n"))
	if _, err := http.Post("http://172.17.0.4:8200/intake/v2/events", "application/x-ndjson", buf); err != nil {
		log.Fatal(err)
	}
}
