package main

import (
	"fmt"
	"log"
	"net/http"
)

func serve(root string, port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", makeHandler(root))
	mux.HandleFunc("/_static/htmx.min.js", serveHtmx)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("serving %s on http://localhost%s", root, addr)
	return http.ListenAndServe(addr, recoverMiddleware(logRequests(mux)))
}
