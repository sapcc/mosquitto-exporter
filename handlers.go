package main

import "net/http"

/*
 * Root and Healthcheck
 */

func serveVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("Mosquitto exporter " + versionString()))
}
