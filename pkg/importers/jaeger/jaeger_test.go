package jaeger

import (
	"fmt"
	apmutil "github.com/justinbarrick/apm-gateway/pkg/apm"
	"github.com/stretchr/testify/assert"
	apmmodel "go.elastic.co/apm/model"
	jaegermodel "github.com/jaegertracing/jaeger/thrift-gen/jaeger"
	"time"
	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

type mockJaeger struct {
	exporter *jaeger.Exporter
}

func newJaeger(url string) *mockJaeger {
	mock := &mockJaeger{}
	mock.register(url)
	return mock
}

func (m *mockJaeger) register(url string) {
	m.exporter, _ = jaeger.NewExporter(jaeger.Options{
		CollectorEndpoint: url,
		Process: jaeger.Process{
			ServiceName: "example-server",
		},
	})
	trace.RegisterExporter(m.exporter)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
}

func (m *mockJaeger) unregister() {
	trace.UnregisterExporter(m.exporter)
}

func (m *mockJaeger) client() *http.Client {
	return &http.Client{
		Transport: &ochttp.Transport{},
	}
}

func TestJaeger(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "")
	}))
	defer ts.Close()

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer wg.Done()

		fmt.Println(r.URL.String())
		spans, err := decodeJaeger(r.Body)
		assert.Nil(t, err)

		for _, span := range spans.Spans {
			s := toAPM(span)
			tags := tagsToMap(span.Tags)
			sampled := true
			assert.Equal(t, s.ID, apmutil.SpanId(uint64(span.SpanId)))
			assert.Equal(t, s.TraceID, apmutil.TraceId(uint64(span.TraceIdHigh), uint64(span.TraceIdLow)))
			assert.Equal(t, s.ParentID, apmutil.SpanId(uint64(span.ParentSpanId)))
			assert.Equal(t, s.Name, span.OperationName)
			assert.Equal(t, s.Timestamp, apmmodel.Time(time.Unix(0, span.StartTime)))
			assert.Equal(t, s.Duration, float64(span.Duration))
			assert.Equal(t, s.Context.Request.URL, apmutil.TagsToURL(tags))
			assert.Equal(t, s.Context.Request.Method, "GET")
			assert.Equal(t, s.Context.Request.Headers[0], apmmodel.Header{
				Key:    "User-Agent",
				Values: []string{tags["http.user_agent"]},
			})
			assert.Equal(t, s.Context.Response.StatusCode, 200)
			assert.ElementsMatch(t, s.Context.Tags, apmutil.TagsToAPM(tags))
			assert.Equal(t, s.Sampled, &sampled)
			assert.ElementsMatch(t, s.Context.Tags, apmutil.TagsToAPM(tags))
			assert.Equal(t, s.SpanCount, apmmodel.SpanCount{})
		}

		fmt.Fprintln(w, "")
	}))
	defer api.Close()

	m := newJaeger(api.URL)
	defer m.unregister()

	resp, err := m.client().Get(ts.URL)
	assert.Nil(t, err)
	resp.Body.Close()

	wg.Wait()
}

func TestTagsToMap(t *testing.T) {
	tags := []*jaegermodel.Tag{}

	host := "example.com:8080"
	status := int64(200)

	tags = append(tags, &jaegermodel.Tag{
		Key: "http.host",
		VStr: &host,
	})

	tags = append(tags, &jaegermodel.Tag{
		VType: jaegermodel.TagType_LONG,
		Key: "http.status",
		VLong: &status,
	})


	assert.Equal(t, tagsToMap(tags), map[string]string{
		"http.host": "example.com:8080",
		"http.status": "200",
	})
}
