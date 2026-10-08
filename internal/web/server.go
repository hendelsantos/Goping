// Package web expõe a API HTTP e os arquivos do dashboard.
package web

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/hendelsantos/Goping/internal/monitor"
)

const maxRequestBody = 1 << 20

type monitorStore interface {
	List() []monitor.Monitor
	Activities() []monitor.Activity
	Create(monitor.CreateInput) (monitor.Monitor, error)
	Delete(string) error
	ClearActivities() error
}

type Server struct {
	store  monitorStore
	logger *slog.Logger
	static http.Handler
}

func NewServer(store monitorStore, logger *slog.Logger, staticDirectory string) http.Handler {
	server := &Server{
		store:  store,
		logger: logger,
		static: http.FileServer(http.Dir(staticDirectory)),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", server.health)
	mux.HandleFunc("/api/v1/monitors", server.monitors)
	mux.HandleFunc("/api/v1/monitors/", server.monitorByID)
	mux.HandleFunc("/api/v1/activities", server.activities)
	mux.HandleFunc("/", server.serveStatic)

	return server.requestLogger(server.securityHeaders(mux))
}

func (s *Server) serveStatic(response http.ResponseWriter, request *http.Request) {
	cleanPath := path.Clean(request.URL.Path)
	allowedFile := cleanPath == "/" || cleanPath == "/index.html" || cleanPath == "/styles.css" || cleanPath == "/app.js" || cleanPath == "/robots.txt" || cleanPath == "/sitemap.xml"
	allowedAsset := strings.HasPrefix(cleanPath, "/assets/") && cleanPath == request.URL.Path
	if !allowedFile && !allowedAsset {
		http.NotFound(response, request)
		return
	}

	s.static.ServeHTTP(response, request)
}

func (s *Server) health(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		methodNotAllowed(response, http.MethodGet)
		return
	}

	writeJSON(response, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "goping",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) monitors(response http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		writeJSON(response, http.StatusOK, map[string]any{"data": s.store.List()})
	case http.MethodPost:
		var input monitor.CreateInput
		if err := decodeJSON(response, request, &input); err != nil {
			writeError(response, http.StatusBadRequest, "invalid_request", err.Error())
			return
		}

		created, err := s.store.Create(input)
		if err != nil {
			writeError(response, http.StatusUnprocessableEntity, "validation_error", err.Error())
			return
		}
		writeJSON(response, http.StatusCreated, map[string]any{"data": created})
	default:
		methodNotAllowed(response, http.MethodGet, http.MethodPost)
	}
}

func (s *Server) monitorByID(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodDelete {
		methodNotAllowed(response, http.MethodDelete)
		return
	}

	id := strings.TrimPrefix(request.URL.Path, "/api/v1/monitors/")
	if id == "" || strings.Contains(id, "/") {
		writeError(response, http.StatusNotFound, "not_found", "monitor não encontrado")
		return
	}

	if err := s.store.Delete(id); err != nil {
		if errors.Is(err, monitor.ErrNotFound) {
			writeError(response, http.StatusNotFound, "not_found", err.Error())
			return
		}
		s.logger.Error("não foi possível excluir monitor", "error", err, "monitor_id", id)
		writeError(response, http.StatusInternalServerError, "internal_error", "não foi possível excluir o monitor")
		return
	}

	response.WriteHeader(http.StatusNoContent)
}

func (s *Server) activities(response http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		writeJSON(response, http.StatusOK, map[string]any{"data": s.store.Activities()})
	case http.MethodDelete:
		if err := s.store.ClearActivities(); err != nil {
			s.logger.Error("não foi possível limpar atividades", "error", err)
			writeError(response, http.StatusInternalServerError, "internal_error", "não foi possível limpar as atividades")
			return
		}
		response.WriteHeader(http.StatusNoContent)
	default:
		methodNotAllowed(response, http.MethodGet, http.MethodDelete)
	}
}

func decodeJSON(response http.ResponseWriter, request *http.Request, destination any) error {
	request.Body = http.MaxBytesReader(response, request.Body, maxRequestBody)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		return errors.New("corpo JSON inválido")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("o corpo deve conter um único objeto JSON")
	}
	return nil
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	if err := json.NewEncoder(response).Encode(value); err != nil {
		slog.Error("não foi possível escrever resposta JSON", "error", err)
	}
}

func writeError(response http.ResponseWriter, status int, code, message string) {
	writeJSON(response, status, map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}

func methodNotAllowed(response http.ResponseWriter, methods ...string) {
	response.Header().Set("Allow", strings.Join(methods, ", "))
	writeError(response, http.StatusMethodNotAllowed, "method_not_allowed", "método HTTP não permitido")
}

func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("X-Content-Type-Options", "nosniff")
		response.Header().Set("X-Frame-Options", "DENY")
		response.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		response.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(response, request)
	})
}

func (s *Server) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		startedAt := time.Now()
		next.ServeHTTP(response, request)
		s.logger.Info("requisição HTTP",
			"method", request.Method,
			"path", request.URL.Path,
			"duration_ms", time.Since(startedAt).Milliseconds(),
		)
	})
}
