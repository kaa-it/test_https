package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/crypto/acme/autocert"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there!")
}

func redirectToHttps(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.Host, ":")
	url := fmt.Sprintf("https://%s:%d", parts[0], 8085) + r.RequestURI
	fmt.Printf("redirect to: %s\n", url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func main() {
	var mode string
	flag.StringVar(&mode, "mode", "production", "run in given mode")
	flag.Parse()

	if mode == "debug" {
		fmt.Println("Running in debug mode")
	} else {
		fmt.Println("Running in production mode")
	}

	if mode == "debug" {
		http.HandleFunc("/", handler)
		go http.ListenAndServe(":8084", http.HandlerFunc(redirectToHttps))
		http.ListenAndServeTLS(":8085", "cert.pem", "key.pem", nil)
	} else {
		certManager := autocert.Manager{
			Prompt: autocert.AcceptTOS,
			Cache:  autocert.DirCache("./certs"),
		}

		srv := &http.Server{
			Addr:      ":443",
			Handler:   http.HandlerFunc(handler),
			TLSConfig: certManager.TLSConfig(),
		}

		go http.ListenAndServe(":80", certManager.HTTPHandler(nil))
		srv.ListenAndServeTLS("", "")
	}
}
