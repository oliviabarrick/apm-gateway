package zipkin

import (
	"encoding/json"
	"fmt"
	apmutil "github.com/justinbarrick/apm-gateway/pkg/apm"
	"github.com/justinbarrick/apm-gateway/pkg/exporters"
	zipkin "github.com/openzipkin/zipkin-go/model"
	apm "go.elastic.co/apm/model"
	"io"
	"log"
	"net/http"
	"strconv"
)

func idToAPM(zipkinID zipkin.ID) apm.SpanID {
	return apmutil.SpanId(uint64(zipkinID))
}

func parentToAPM(zipkinParent *zipkin.ID) (parentId apm.SpanID) {
	if zipkinParent != nil {
		parentId = idToAPM(*zipkinParent)
	}
	return
}

func traceIdToAPM(zipkinTraceID zipkin.TraceID) apm.TraceID {
	return apmutil.TraceId(zipkinTraceID.High, zipkinTraceID.Low)
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
				URL:    apmutil.TagsToURL(zipkinSpan.Tags),
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
			Tags: apmutil.TagsToAPM(zipkinSpan.Tags),
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

type Importer struct {
	exporter exporter.Exporter
}

func (i *Importer) SetExporter(e exporter.Exporter) {
	i.exporter = e
}

func (i *Importer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	spans, err := decodeZipkin(r.Body)
	if err != nil {
		log.Println(err)
	}

	for _, span := range spans {
		if err := i.exporter.SendToAPM(toAPM(span)); err != nil {
			log.Println(err)
		}
	}

	fmt.Fprintf(w, "Hello, %q", r.URL.Path)
}
