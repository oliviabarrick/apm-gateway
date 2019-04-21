package zipkin

import (
	"fmt"
	apmutil "github.com/justinbarrick/apm-gateway/pkg/apm"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinmodel "github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/reporter"
	reporterHttp "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/stretchr/testify/assert"
	apmmodel "go.elastic.co/apm/model"
	"go.opencensus.io/exporter/zipkin"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

type mockZipkin struct {
	reporter reporter.Reporter
	exporter *zipkin.Exporter
}

func newZipkin(url string) *mockZipkin {
	mock := &mockZipkin{}
	mock.register(url)
	return mock
}

func (m *mockZipkin) register(url string) {
	localEndpoint, _ := openzipkin.NewEndpoint("example-server", "192.168.1.5:5454")

	m.reporter = reporterHttp.NewReporter(url)
	m.exporter = zipkin.NewExporter(m.reporter, localEndpoint)
	trace.RegisterExporter(m.exporter)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
}

func (m *mockZipkin) unregister() {
	trace.UnregisterExporter(m.exporter)
	m.reporter.Close()
}

func (m *mockZipkin) client() *http.Client {
	return &http.Client{
		Transport: &ochttp.Transport{},
	}
}

func TestZipkin(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "")
	}))
	defer ts.Close()

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer wg.Done()

		spans, err := decodeZipkin(r.Body)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(spans))

		for _, span := range spans {
			apm := toAPM(span)
			assert.Equal(t, idToAPM(span.SpanContext.ID), apm.ID)
			assert.Equal(t, traceIdToAPM(span.SpanContext.TraceID), apm.TraceID)
			assert.Equal(t, apmmodel.SpanID{}, apm.ParentID)
			assert.Equal(t, span.Name, apm.Name)
			assert.Equal(t, apm.Type, string(span.Kind))
			assert.Equal(t, apm.Timestamp, apmmodel.Time(span.Timestamp))
			assert.Equal(t, apm.Duration, float64(span.Duration.Nanoseconds()/1000000.0))
			assert.Equal(t, apm.Result, "200")
			assert.Equal(t, apm.Context.Request.URL, apmutil.TagsToURL(span.Tags))
			assert.Equal(t, apm.Context.Request.Method, "GET")
			assert.Equal(t, apm.Context.Request.Headers[0], apmmodel.Header{
				Key:    "User-Agent",
				Values: []string{span.Tags["http.user_agent"]},
			})
			assert.Equal(t, apm.Context.Request.Socket, clientToAPM(span.RemoteEndpoint))
			assert.Equal(t, apm.Context.Service, serviceToAPM(span.LocalEndpoint))
			assert.Equal(t, apm.Context.Response.StatusCode, 200)
			assert.ElementsMatch(t, apm.Context.Tags, apmutil.TagsToAPM(span.Tags))
			assert.Equal(t, apm.Sampled, span.SpanContext.Sampled)
			assert.Equal(t, apm.SpanCount, apmmodel.SpanCount{})
		}

		fmt.Fprintln(w, "")
	}))
	defer api.Close()

	m := newZipkin(api.URL)
	defer m.unregister()

	resp, err := m.client().Get(ts.URL)
	assert.Nil(t, err)
	resp.Body.Close()

	wg.Wait()
}

func TestZipkinIDToAPM(t *testing.T) {
	assert.Equal(t, idToAPM(zipkinmodel.ID(0x1234567890123456)), apmmodel.SpanID([8]byte{
		0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56,
	}))
}

func TestZipkinParentToAPM(t *testing.T) {
	assert.Equal(t, parentToAPM(nil), apmmodel.SpanID{})

	parentId := zipkinmodel.ID(0x1234567890123456)
	assert.Equal(t, parentToAPM(&parentId), apmmodel.SpanID([8]byte{
		0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56,
	}))
}

func TestZipkinTraceIDToAPM(t *testing.T) {
	assert.Equal(t, traceIdToAPM(zipkinmodel.TraceID{
		High: 0x1234567890123456,
		Low:  0x3456028234520945,
	}), apmmodel.TraceID([16]byte{
		0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56,
		0x34, 0x56, 0x02, 0x82, 0x34, 0x52, 0x09, 0x45,
	}))
}

func TestZipkinClientToAPM(t *testing.T) {
	assert.Nil(t, clientToAPM(nil))

	assert.Equal(t, clientToAPM(&zipkinmodel.Endpoint{
		IPv4: net.IPv4(12, 23, 34, 11),
		Port: 1234,
	}), &apmmodel.RequestSocket{
		RemoteAddress: "12.23.34.11:1234",
	})

	assert.Equal(t, clientToAPM(&zipkinmodel.Endpoint{
		IPv4: net.IPv4(12, 23, 34, 11),
		Port: 0,
	}), &apmmodel.RequestSocket{
		RemoteAddress: "12.23.34.11",
	})

	assert.Equal(t, clientToAPM(&zipkinmodel.Endpoint{
		IPv6: net.ParseIP("2001:db8::68"),
		Port: 1234,
	}), &apmmodel.RequestSocket{
		RemoteAddress: "[2001:db8::68]:1234",
	})
}

func TestZipkinServiceToAPM(t *testing.T) {
	assert.Nil(t, serviceToAPM(nil))
	assert.Equal(t, serviceToAPM(&zipkinmodel.Endpoint{
		ServiceName: "hello",
	}), &apmmodel.Service{
		Name: "hello",
	})
}
