package web

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/hendelsantos/Goping/internal/monitor"
)

func TestHealthEndpoint(t *testing.T) {
	t.Parallel()

	handler := newTestServer(t)
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("GET /healthz status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("GET /healthz missing security headers")
	}
}

func TestMonitorLifecycle(t *testing.T) {
	t.Parallel()

	handler := newTestServer(t)
	requestBody := []byte(`{
		"name":"API pública",
		"url":"https://example.com/health",
		"interval_seconds":300,
		"timeout_seconds":10
	}`)

	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/monitors", bytes.NewReader(requestBody))
	createResponse := httptest.NewRecorder()
	handler.ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("POST /api/v1/monitors status = %d, want %d; body = %s", createResponse.Code, http.StatusCreated, createResponse.Body.String())
	}

	var created struct {
		Data monitor.Monitor `json:"data"`
	}
	if err := json.NewDecoder(createResponse.Body).Decode(&created); err != nil {
		t.Fatalf("decode created monitor: %v", err)
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/api/v1/monitors", nil)
	listResponse := httptest.NewRecorder()
	handler.ServeHTTP(listResponse, listRequest)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/monitors status = %d, want %d", listResponse.Code, http.StatusOK)
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/monitors/"+created.Data.ID, nil)
	deleteResponse := httptest.NewRecorder()
	handler.ServeHTTP(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("DELETE /api/v1/monitors/{id} status = %d, want %d", deleteResponse.Code, http.StatusNoContent)
	}
}

func TestCreateMonitorRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	handler := newTestServer(t)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/monitors", bytes.NewBufferString(`{
		"name":"Rede interna",
		"url":"file:///etc/passwd",
		"interval_seconds":300,
		"timeout_seconds":10
	}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("POST invalid monitor status = %d, want %d", response.Code, http.StatusUnprocessableEntity)
	}
}

func TestStaticHandlerDoesNotExposeBackendFiles(t *testing.T) {
	t.Parallel()

	handler := newTestServer(t)
	request := httptest.NewRequest(http.MethodGet, "/go.mod", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("GET /go.mod status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func newTestServer(t *testing.T) http.Handler {
	t.Helper()

	store, err := monitor.NewStore(filepath.Join(t.TempDir(), "goping.json"))
	if err != nil {
		t.Fatalf("monitor.NewStore() error = %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewServer(store, logger, t.TempDir())
}
