package main

import (
	"context"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var httpServer http.Server
var outChan chan<- byte

var (
	LedControl = regexp.MustCompile(`^/led/(.*)$`)
)

func restApi(outChanNew chan<- byte, wg *sync.WaitGroup) {
	defer wg.Done()

	outChan = outChanNew

	mux := http.NewServeMux()

	httpServer = http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	mux.Handle("/", &homeHandler{})
	mux.Handle("/led/", &ledHandler{})

	httpServer.ListenAndServe()
	log.Printf("Rest API http done\n")

}

func shutdownRestApi() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Printf("Shutting down rest API")
	var err = httpServer.Shutdown(ctx)
	if err != nil {
		log.Printf("HTTP server shutdown error: %s\n", err)
	}
	log.Printf("Finished shutting down rest API\n")

}

type ledHandler struct{}

func (h *ledHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("ledHandler\n")

	switch {
	case r.Method == http.MethodGet:
		handleLedPost(w, r)

	}
}

func handleLedPost(w http.ResponseWriter, r *http.Request) {
	log.Printf("handleLedPost\n")

	var matches = LedControl.FindStringSubmatch(r.URL.Path)
	if len(matches) > 0 {
		var ledNumberString = matches[1]
		ledNumber, err := strconv.Atoi(ledNumberString)
		if err != nil {
			// ... handle error
			panic(err)
		}
		outChan <- byte(ledNumber)
	}
}

type homeHandler struct{}

func (h *homeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is my new home page"))
}
