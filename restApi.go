package main

import (
	"context"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var httpServer http.Server

var (
	LedControl = regexp.MustCompile(`^/led/(.*)$`)
)

type RestAPI struct {
	logger    *zerolog.Logger
	inChan    chan byte
	outChan   chan byte
	waitGroup *sync.WaitGroup
}

func NewRestApi(logger *zerolog.Logger, inChan chan byte, outChan chan byte, waitGroup *sync.WaitGroup) *RestAPI {
	return &RestAPI{
		logger:    ptr(logger.With().Str(LogKey.Module, "RestAPI").Logger()),
		inChan:    inChan,
		outChan:   outChan,
		waitGroup: waitGroup,
	}
}

func (api *RestAPI) Start(ctx context.Context) {
	api.waitGroup.Add(1)

	go api.listen()
	go api.waitForCancel(ctx)
}

func (api *RestAPI) listen() {
	defer api.waitGroup.Done()

	mux := http.NewServeMux()

	httpServer = http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	mux.Handle("/", &homeHandler{
		logger: api.logger,
	})
	mux.Handle("/led/", &ledHandler{
		logger:  api.logger,
		outChan: api.outChan,
	})
	api.logger.Info().Msgf("Listening")

	err := httpServer.ListenAndServe()
	api.logger.Info().Msgf("Done: %v", err)
}

func (api *RestAPI) waitForCancel(ctx context.Context) {
	select {
	case <-ctx.Done():
		{
			api.logger.Info().Msg("Stopping")
			api.stop()
			return
		}
	}
}

func (api *RestAPI) stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	api.logger.Info().Msg("Shutting down")
	var err = httpServer.Shutdown(ctx)
	if err != nil {
		api.logger.Error().Msgf("HTTP server shutdown error: %s", err)
	}
	api.logger.Info().Msg("Finished shutting down")
}

type ledHandler struct {
	logger  *zerolog.Logger
	outChan chan byte
}

func (lh *ledHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lh.logger.Info().Msg("ledHandler")

	switch {
	case r.Method == http.MethodGet:
		lh.handleLedPost(w, r)
	}
}

func (lh *ledHandler) handleLedPost(w http.ResponseWriter, r *http.Request) {
	lh.logger.Info().Msg("handleLedPost")

	var matches = LedControl.FindStringSubmatch(r.URL.Path)
	if len(matches) > 0 {
		var ledNumberString = matches[1]
		ledNumber, err := strconv.Atoi(ledNumberString)
		if err != nil {
			// ... handle error
			panic(err)
		}
		lh.outChan <- byte(ledNumber)
	}
}

type homeHandler struct {
	logger *zerolog.Logger
}

func (hh *homeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("ButtonBox Home Page"))
	if err != nil {
		hh.logger.Error().Msgf("Write failed: %v", err)
		return
	}
}
