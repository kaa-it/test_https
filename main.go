package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
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
		if _, err := os.Stat("./certs"); os.IsNotExist(err) {
			if err := os.MkdirAll("./certs", os.ModePerm); err != nil {
				panic(err)
			}
		}

		certManager := autocert.Manager{
			Prompt: autocert.AcceptTOS,
			Cache:  autocert.DirCache("./certs"),
		}

		srv := &http.Server{
			Addr:      ":8443",
			Handler:   http.HandlerFunc(handler),
			TLSConfig: certManager.TLSConfig(),
		}

		go func() {
			if err := http.ListenAndServe(":8080", certManager.HTTPHandler(nil)); err != nil {
				log.Printf("http server: %s", err)
			}
		}()

		if err := srv.ListenAndServeTLS("", ""); err != nil {
			log.Printf("https server: %s", err)
		}
	}
}
