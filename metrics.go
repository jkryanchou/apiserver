package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
)

type MetricsFooKey struct {
	Key1 string
	Key2 string
}

var (
	MetricsFooCounter sync.Map // map[MetricsFooKey]*int64
)

func Metrics(rw http.ResponseWriter, req *http.Request) {
	var w io.Writer = rw
	if strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
		gz := gzip.NewWriter(rw)
		defer gz.Close()
		w = gz
		rw.Header().Set("Content-Encoding", "gzip")
	}

	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)

	io.WriteString(w, "# HELP apiserver_foo_count transmit bytes\n")
	io.WriteString(w, "# TYPE apiserver_foo_count gauge\n")
	MetricsFooCounter.Range(func(key, value interface{}) bool {
		k := key.(MetricsFooKey)
		v := atomic.LoadInt64(value.(*int64))
		fmt.Fprintf(w, "apiserver_foo_count{key1=\"%s\",key2=\"%s\"} %d\n", k.Key1, k.Key2, v)
		return true
	})
}
