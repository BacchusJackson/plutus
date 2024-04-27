package api

import (
	"context"
	"log/slog"
	"net/http"
)

type AgentOptionFunc func(*Agent)

func WithDefault() AgentOptionFunc {
	return func(a *Agent) {
		a.httpClient = http.DefaultClient
	}
}

func WithHTTPClient(c *http.Client) AgentOptionFunc {
	return func(a *Agent) {
		a.httpClient = c
	}
}

type Agent struct {
	httpClient *http.Client
	middleware []middlewareFunc
}

func NewAgent(opts ...AgentOptionFunc) *Agent {
	agent := &Agent{}
	WithDefault()(agent)
	for _, opt := range opts {
		opt(agent)
	}
	return agent
}

func (a *Agent) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	var withMiddleware = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, m := range a.middleware {
			err := m(w, r)
			if err != nil {
				slog.Error("middleware", "error", err)
				return
			}
		}

		mux.ServeHTTP(w, r)
	})

	server := http.Server{
		Handler: withMiddleware,
	}

	return server.ListenAndServe()
}

type middlewareFunc func(http.ResponseWriter, *http.Request) error

func loggingMiddleware(w http.ResponseWriter, r *http.Request) error {
	slog.Debug("request", "request_uri", r.RequestURI, "remote_addr", r.RemoteAddr)
	return nil
}
