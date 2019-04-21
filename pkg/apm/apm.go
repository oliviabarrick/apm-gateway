package apm

import (
	"strings"
	"encoding/binary"
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

func TagsToAPM(inputTags map[string]string) (tags apm.StringMap) {
	for key, value := range inputTags {
		tags = append(tags, apm.StringMapItem{
			Key:   strings.Replace(key, ".", "_", -1),
			Value: value,
		})
	}
	return
}

func IntToBytes(num uint64) [8]byte {
	finalByteSlice := make([]byte, 8)
	finalBytes := [8]byte{}

	binary.BigEndian.PutUint64(finalByteSlice, num)
	copy(finalBytes[:], finalByteSlice)
	return finalBytes
}

func TraceId(high uint64, low uint64) apm.TraceID {
	traceId := [16]byte{}

	highSlice := IntToBytes(high)
	copy(traceId[:8], highSlice[:])

	lowSlice := IntToBytes(low)
	copy(traceId[8:], lowSlice[:])

	return apm.TraceID(traceId)
}

func SpanId(num uint64) apm.SpanID {
	return apm.SpanID(IntToBytes(num))
}

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

func TagsToURL(tags map[string]string) apm.URL {
	return UrlToAPM(url.URL{
		Scheme: "http",
		Host:   tags["http.host"],
		Path:   tags["http.path"],
	})
}
