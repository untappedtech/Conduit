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
}

func NewServer(apiService *service.APIService, serverConfig *domain.ServerConfig, responseEncoder *ResponseEncoder, customBasePath ...string) *Server {
	apiHandler := NewAPIHandler(apiService, serverConfig, responseEncoder)
	serveMux := http.NewServeMux()
	if len(customBasePath) > 0 && customBasePath[0] != "" {
		apiHandler.RegisterRoutes(serveMux, customBasePath[0])
	} else {
		apiHandler.RegisterRoutes(serveMux)
	}

	return &Server{
		apiHandler:   apiHandler,
		serverConfig: serverConfig,
		serveMux:     serveMux,
	}
}

func (serverInstance *Server) BasePath() string {
	return serverInstance.apiHandler.BasePath()
}

// Handler returns the underlying HTTP router.
func (serverInstance *Server) Handler() http.Handler {
	return serverInstance.serveMux
}

// ServeHTTP implements http.Handler for Server.
func (serverInstance *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serverInstance.serveMux.ServeHTTP(w, r)
}

func (serverInstance *Server) ListenAndServe(authChain []domain.AuthProvider, tokenExtractor auth.TokenExtractor, errorResponder auth.ErrorResponder) error {
	wrappedMux := auth.AuthMiddleware(authChain, tokenExtractor, errorResponder)(serverInstance.serveMux)

	serverAddress := fmt.Sprintf("%s:%d", serverInstance.serverConfig.Server.Host, serverInstance.serverConfig.Server.Port)
	serverInstance.httpServer = &http.Server{
		Addr:    serverAddress,
		Handler: wrappedMux,
	}

	return serverInstance.httpServer.ListenAndServe()
}

func (serverInstance *Server) Shutdown(ctx context.Context) error {
	if serverInstance.httpServer != nil {
		return serverInstance.httpServer.Shutdown(ctx)
	}
	return nil
}
