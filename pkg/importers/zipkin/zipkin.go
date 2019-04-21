package zipkin

import (
	"strconv"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"net/http"
	"strings"
	apm "go.elastic.co/apm/model"
	zipkin "github.com/openzipkin/zipkin-go/model"
	apmutil "github.com/justinbarrick/apm-gateway/pkg/apm"
)

func traceIdToAPM(zipkinTraceID zipkin.TraceID) apm.TraceID {
	zipkinByteSlice := make([]byte, 8)
	zipkinBytes := [16]byte{}

	binary.BigEndian.PutUint64(zipkinByteSlice, uint64(zipkinTraceID.High))
	copy(zipkinBytes[:8], zipkinByteSlice)

	binary.BigEndian.PutUint64(zipkinByteSlice, uint64(zipkinTraceID.Low))
	copy(zipkinBytes[8:], zipkinByteSlice)

	return apm.TraceID(zipkinBytes)
}

func idToAPM(zipkinID zipkin.ID) apm.SpanID {
	zipkinByteSlice := make([]byte, 8)
	zipkinBytes := [8]byte{}

	binary.BigEndian.PutUint64(zipkinByteSlice, uint64(zipkinID))
	copy(zipkinBytes[:], zipkinByteSlice)

	return apm.SpanID(zipkinBytes)
}

func tagsToAPM(zipkinTags map[string]string) (tags apm.StringMap) {
	for key, value := range zipkinTags {
		tags = append(tags, apm.StringMapItem{
			Key: strings.Replace(key, ".", "_", -1),
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

	return &apm.RequestSocket{
		RemoteAddress: fmt.Sprintf("%s:%d", addr, zipkinEndpoint.Port),
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

func parentToAPM(zipkinParent *zipkin.ID) (parentId apm.SpanID) {
	if zipkinParent != nil {
		parentId = idToAPM(*zipkinParent)
	}
	return 
}

func urlToURL(zipkinTags map[string]string) url.URL {
	return url.URL{
		Scheme: "http",
		Host: zipkinTags["http.host"],
		Path: zipkinTags["http.path"],
	}
}

func toAPM(zipkinSpan zipkin.SpanModel) *apm.Transaction {
	statusCode, _ := strconv.Atoi(zipkinSpan.Tags["http.status_code"])

	return &apm.Transaction{
		ID: idToAPM(zipkinSpan.SpanContext.ID),
		TraceID: traceIdToAPM(zipkinSpan.SpanContext.TraceID),
		ParentID: parentToAPM(zipkinSpan.SpanContext.ParentID),
		Name: zipkinSpan.Name,
		Type: string(zipkinSpan.Kind),
		Timestamp: apm.Time(zipkinSpan.Timestamp),
		Duration: float64(zipkinSpan.Duration.Nanoseconds() / 1000000.0),
		Result: zipkinSpan.Tags["http.status_code"],
		Context: &apm.Context{
			Request: &apm.Request{
				URL: apmutil.UrlToAPM(urlToURL(zipkinSpan.Tags)),
				Method: zipkinSpan.Tags["http.method"],
				Headers: []apm.Header{
					apm.Header{
						Key: "User-Agent",
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

func Handler(w http.ResponseWriter, r *http.Request) {
	spans := []zipkin.SpanModel{}

	if err := json.NewDecoder(r.Body).Decode(&spans); err != nil {
		log.Println(err)
	}

	for _, span := range spans {
		if err := apmutil.SendToAPM(toAPM(span)); err != nil {
			log.Println(err)
		}
	}

	fmt.Fprintf(w, "Hello, %q", r.URL.Path)
}
