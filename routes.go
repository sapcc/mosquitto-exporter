package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/prometheus/client_golang/prometheus"
)

type route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type routes []route

var routesDefinition = routes{
	route{
		"Root",
		"GET",
		"/",
		serveVersion,
	},
	route{
		"Healthcheck",
		"GET",
		"/healthcheck",
		serveVersion,
	},
}

func newRouter() *mux.Router {
	middlewareChain := alice.New(loggingHandler, combineLogHandler, servedByHandler)

	router := mux.NewRouter().StrictSlash(true)
	for _, r := range routesDefinition {
		router.
			Methods(r.Method).
			Path(r.Pattern).
			Name(r.Name).
			Handler(middlewareChain.Then(prometheus.InstrumentHandler(r.Name, r.HandlerFunc)))
	}

	// add metrics
	router.
		Methods("GET").
		Path("/metrics").
		Name("Metrics").
		Handler(middlewareChain.Then(prometheus.Handler()))

	return router
}
