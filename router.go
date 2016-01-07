package main

import (
	"net/http"
	"regexp"
)

type route struct {
	pattern *regexp.Regexp
	method  string
	handler http.Handler
}

type MyHandler struct {
	routes []*route
}

func (h *MyHandler) HandleStatic(r string, method string, handler http.Handler) {
	re := regexp.MustCompile(r)
	h.routes = append(h.routes, &route{re, method, handler})
}

func (h *MyHandler) HandleFunc(r string, v string, handler func(http.ResponseWriter, *http.Request)) {
	re := regexp.MustCompile(r)
	h.routes = append(h.routes, &route{re, v, http.HandlerFunc(handler)})
}

func (h *MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range h.routes {
		if route.pattern.MatchString(r.URL.Path) && route.method == r.Method {
			route.handler.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}