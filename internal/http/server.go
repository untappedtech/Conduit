package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/untappedtech/conduit/internal/auth"
	"github.com/untappedtech/conduit/internal/domain"
	"github.com/untappedtech/conduit/internal/service"
)

type Server struct {
	apiHandler   *APIHandler
	serverConfig *domain.ServerConfig
	serveMux     *http.ServeMux
	httpServer   *http.Server
	rateLimiter  *RateLimiter
}

func NewServer(apiService *service.APIService, serverConfig *domain.ServerConfig, responseEncoder *ResponseEncoder, customBasePath ...string) *Server {
	apiHandler := NewAPIHandler(apiService, serverConfig, responseEncoder)
	serveMux := http.NewServeMux()
	if len(customBasePath) > 0 && customBasePath[0] != "" {
		apiHandler.RegisterRoutes(serveMux, customBasePath[0])
	} else {
		apiHandler.RegisterRoutes(serveMux)
	}

	var rl *RateLimiter
	if serverConfig != nil && serverConfig.RateLimit.Enabled {
		rl = NewRateLimiter(serverConfig.RateLimit, responseEncoder)
	}

	return &Server{
		apiHandler:   apiHandler,
		serverConfig: serverConfig,
		serveMux:     serveMux,
		rateLimiter:  rl,
	}
}

func (serverInstance *Server) BasePath() string {
	return serverInstance.apiHandler.BasePath()
}

// Handler returns the underlying HTTP router.
func (serverInstance *Server) Handler() http.Handler {
	return serverInstance.serveMux
}

// RateLimiter returns the active rate limiter, if configured.
func (serverInstance *Server) RateLimiter() *RateLimiter {
	return serverInstance.rateLimiter
}

// ServeHTTP implements http.Handler for Server.
func (serverInstance *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serverInstance.serveMux.ServeHTTP(w, r)
}

func (serverInstance *Server) ListenAndServe(authChain []domain.AuthProvider, tokenExtractor auth.TokenExtractor, errorResponder auth.ErrorResponder) error {
	var handler http.Handler = serverInstance.serveMux
	handler = auth.AuthMiddleware(authChain, tokenExtractor, errorResponder)(handler)

	if serverInstance.serverConfig != nil && serverInstance.serverConfig.RateLimit.Enabled {
		if serverInstance.rateLimiter == nil {
			serverInstance.rateLimiter = NewRateLimiter(serverInstance.serverConfig.RateLimit, serverInstance.apiHandler.responseEncoder, tokenExtractor)
		}
		handler = serverInstance.rateLimiter.Middleware()(handler)
	}

	serverAddress := fmt.Sprintf("%s:%d", serverInstance.serverConfig.Server.Host, serverInstance.serverConfig.Server.Port)
	serverInstance.httpServer = &http.Server{
		Addr:    serverAddress,
		Handler: handler,
	}

	return serverInstance.httpServer.ListenAndServe()
}

func (serverInstance *Server) Shutdown(ctx context.Context) error {
	if serverInstance.rateLimiter != nil {
		serverInstance.rateLimiter.Close()
	}
	if serverInstance.httpServer != nil {
		return serverInstance.httpServer.Shutdown(ctx)
	}
	return nil
}
