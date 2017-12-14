package main

import (
	"fmt"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `Ipinfo lookup:

Usage:
    curl -v http://%s/ipinfo/127.0.0.1

`, r.Host)
}
