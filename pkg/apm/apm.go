package apm

import (
	"encoding/binary"
	apm "go.elastic.co/apm/model"
	"net/url"
	"strings"
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
