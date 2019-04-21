package zipkin

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	apmutil "github.com/justinbarrick/apm-gateway/pkg/apm"
	zipkin "github.com/openzipkin/zipkin-go/model"
	apm "go.elastic.co/apm/model"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func idToAPM(zipkinID zipkin.ID) apm.SpanID {
	zipkinByteSlice := make([]byte, 8)
	zipkinBytes := [8]byte{}

	binary.BigEndian.PutUint64(zipkinByteSlice, uint64(zipkinID))
	copy(zipkinBytes[:], zipkinByteSlice)

	return apm.SpanID(zipkinBytes)
}

func parentToAPM(zipkinParent *zipkin.ID) (parentId apm.SpanID) {
	if zipkinParent != nil {
		parentId = idToAPM(*zipkinParent)
	}
	return
}

func traceIdToAPM(zipkinTraceID zipkin.TraceID) apm.TraceID {
	zipkinByteSlice := make([]byte, 8)
	zipkinBytes := [16]byte{}

	binary.BigEndian.PutUint64(zipkinByteSlice, uint64(zipkinTraceID.High))
	copy(zipkinBytes[:8], zipkinByteSlice)

	binary.BigEndian.PutUint64(zipkinByteSlice, uint64(zipkinTraceID.Low))
	copy(zipkinBytes[8:], zipkinByteSlice)

	return apm.TraceID(zipkinBytes)
}

func tagsToAPM(zipkinTags map[string]string) (tags apm.StringMap) {
	for key, value := range zipkinTags {
		tags = append(tags, apm.StringMapItem{
			Key:   strings.Replace(key, ".", "_", -1),
			Value: value,
		})
	}
	return
}

func clientToAPM(zipkinEndpoint *zipkin.Endpoint) *apm.RequestSocket {
	if zipkinEndpoint == nil {
		return nil
	}

	addr := zipkinEndpoint.IPv4.String()

	if len(zipkinEndpoint.IPv6) != 0 {
		addr = fmt.Sprintf("[%s]", zipkinEndpoint.IPv6.String())
	}

	if zipkinEndpoint.Port != 0 {
		addr = fmt.Sprintf("%s:%d", addr, zipkinEndpoint.Port)
	}

	return &apm.RequestSocket{
		RemoteAddress: addr,
	}
}

func serviceToAPM(zipkinEndpoint *zipkin.Endpoint) *apm.Service {
	if zipkinEndpoint == nil {
		return nil
	}

	return &apm.Service{
		Name: zipkinEndpoint.ServiceName,
	}
}

func urlToURL(zipkinTags map[string]string) url.URL {
	return url.URL{
		Scheme: "http",
		Host:   zipkinTags["http.host"],
		Path:   zipkinTags["http.path"],
	}
}

func toAPM(zipkinSpan zipkin.SpanModel) *apm.Transaction {
	statusCode, _ := strconv.Atoi(zipkinSpan.Tags["http.status_code"])

	return &apm.Transaction{
		ID:        idToAPM(zipkinSpan.SpanContext.ID),
		TraceID:   traceIdToAPM(zipkinSpan.SpanContext.TraceID),
		ParentID:  parentToAPM(zipkinSpan.SpanContext.ParentID),
		Name:      zipkinSpan.Name,
		Type:      string(zipkinSpan.Kind),
		Timestamp: apm.Time(zipkinSpan.Timestamp),
		Duration:  float64(zipkinSpan.Duration.Nanoseconds() / 1000000.0),
		Result:    zipkinSpan.Tags["http.status_code"],
		Context: &apm.Context{
			Request: &apm.Request{
				URL:    apmutil.UrlToAPM(urlToURL(zipkinSpan.Tags)),
				Method: zipkinSpan.Tags["http.method"],
				Headers: []apm.Header{
					{
						Key:    "User-Agent",
						Values: []string{zipkinSpan.Tags["http.user_agent"]},
					},
				},
				Socket: clientToAPM(zipkinSpan.RemoteEndpoint),
			},
			Service: serviceToAPM(zipkinSpan.LocalEndpoint),
			Response: &apm.Response{
				StatusCode: statusCode,
			},
			Tags: tagsToAPM(zipkinSpan.Tags),
		},
		Sampled: zipkinSpan.SpanContext.Sampled,
		SpanCount: apm.SpanCount{
			Dropped: 0,
			Started: 0,
		},
	}
}

func decodeZipkin(body io.Reader) (spans []zipkin.SpanModel, err error) {
	return spans, json.NewDecoder(body).Decode(&spans)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	spans, err := decodeZipkin(r.Body)
	if err != nil {
		log.Println(err)
	}

	for _, span := range spans {
		if err := apmutil.SendToAPM(toAPM(span)); err != nil {
			log.Println(err)
		}
	}

	fmt.Fprintf(w, "Hello, %q", r.URL.Path)
}
