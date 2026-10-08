// Package monitor concentra os dados e as regras básicas dos monitores.
package monitor

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var ErrNotFound = errors.New("monitor não encontrado")

type Monitor struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	URL             string    `json:"url"`
	IntervalSeconds int       `json:"interval_seconds"`
	TimeoutSeconds  int       `json:"timeout_seconds"`
	Enabled         bool      `json:"enabled"`
	Status          string    `json:"status"`
	Uptime          float64   `json:"uptime"`
	LatencyMS       *int64    `json:"latency_ms"`
	LastCheck       string    `json:"last_check"`
	CreatedAt       time.Time `json:"created_at"`
}

type Activity struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Time        string `json:"time"`
}

type CreateInput struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	IntervalSeconds int    `json:"interval_seconds"`
	TimeoutSeconds  int    `json:"timeout_seconds"`
}

type persistedData struct {
	Monitors   []Monitor  `json:"monitors"`
	Activities []Activity `json:"activities"`
}

type Store struct {
	mu         sync.RWMutex
	path       string
	monitors   []Monitor
	activities []Activity
}

func NewStore(path string) (*Store, error) {
	store := &Store{path: path}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Store) List() []Monitor {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return append([]Monitor(nil), s.monitors...)
}

func (s *Store) Activities() []Activity {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return append([]Activity(nil), s.activities...)
}

func (s *Store) Create(input CreateInput) (Monitor, error) {
	if err := validateInput(input); err != nil {
		return Monitor{}, err
	}

	id, err := newID("mon")
	if err != nil {
		return Monitor{}, fmt.Errorf("gerar identificador: %w", err)
	}

	created := Monitor{
		ID:              id,
		Name:            strings.TrimSpace(input.Name),
		URL:             strings.TrimSpace(input.URL),
		IntervalSeconds: input.IntervalSeconds,
		TimeoutSeconds:  input.TimeoutSeconds,
		Enabled:         true,
		Status:          "unknown",
		Uptime:          0,
		LastCheck:       "Aguardando",
		CreatedAt:       time.Now().UTC(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.monitors = append([]Monitor{created}, s.monitors...)
	s.activities = append([]Activity{{
		ID:          "activity-" + id,
		Type:        "info",
		Title:       created.Name + " foi adicionado",
		Description: fmt.Sprintf("Primeira verificação agendada a cada %d minutos.", input.IntervalSeconds/60),
		Time:        "agora",
	}}, s.activities...)
	if err := s.saveLocked(); err != nil {
		return Monitor{}, err
	}

	return created, nil
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, item := range s.monitors {
		if item.ID != id {
			continue
		}

		s.monitors = append(s.monitors[:index], s.monitors[index+1:]...)
		s.activities = append([]Activity{{
			ID:          fmt.Sprintf("removed-%s-%d", id, time.Now().UnixNano()),
			Type:        "info",
			Title:       item.Name + " foi removido",
			Description: "O monitor foi removido do GoPing.",
			Time:        "agora",
		}}, s.activities...)
		return s.saveLocked()
	}

	return ErrNotFound
}

func (s *Store) ClearActivities() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.activities = []Activity{}
	return s.saveLocked()
}

func validateInput(input CreateInput) error {
	name := strings.TrimSpace(input.Name)
	if name == "" || len(name) > 80 {
		return errors.New("nome deve ter entre 1 e 80 caracteres")
	}

	parsedURL, err := url.ParseRequestURI(strings.TrimSpace(input.URL))
	if err != nil || parsedURL.Host == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return errors.New("URL deve usar HTTP ou HTTPS")
	}
	if parsedURL.User != nil {
		return errors.New("URL não pode conter credenciais")
	}
	if input.IntervalSeconds < 300 || input.IntervalSeconds > 86400 {
		return errors.New("intervalo deve ficar entre 300 e 86400 segundos")
	}
	if input.TimeoutSeconds < 1 || input.TimeoutSeconds > 30 {
		return errors.New("timeout deve ficar entre 1 e 30 segundos")
	}

	return nil
}

func (s *Store) load() error {
	content, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		s.monitors, s.activities = seedData()
		return nil
	}
	if err != nil {
		return fmt.Errorf("ler dados: %w", err)
	}

	var data persistedData
	if err := json.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("decodificar dados: %w", err)
	}
	s.monitors = data.Monitors
	s.activities = data.Activities
	return nil
}

// A troca de nome torna a gravação atômica: o arquivo anterior permanece válido em caso de falha.
func (s *Store) saveLocked() error {
	content, err := json.MarshalIndent(persistedData{Monitors: s.monitors, Activities: s.activities}, "", "  ")
	if err != nil {
		return fmt.Errorf("codificar dados: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("criar diretório de dados: %w", err)
	}

	temporaryPath := s.path + ".tmp"
	if err := os.WriteFile(temporaryPath, content, 0o600); err != nil {
		return fmt.Errorf("gravar dados temporários: %w", err)
	}
	if err := os.Rename(temporaryPath, s.path); err != nil {
		return fmt.Errorf("publicar dados: %w", err)
	}
	return nil
}

func newID(prefix string) (string, error) {
	buffer := make([]byte, 8)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return prefix + "_" + hex.EncodeToString(buffer), nil
}

func seedData() ([]Monitor, []Activity) {
	now := time.Now().UTC()
	latencies := []int64{142, 198, 211}
	monitors := []Monitor{
		{ID: "website", Name: "Site institucional", URL: "https://goping.dev", IntervalSeconds: 300, TimeoutSeconds: 10, Enabled: true, Status: "up", Uptime: 99.99, LatencyMS: &latencies[0], LastCheck: "Agora", CreatedAt: now},
		{ID: "payments-api", Name: "API de pagamentos", URL: "https://api.goping.dev/payments", IntervalSeconds: 300, TimeoutSeconds: 10, Enabled: true, Status: "up", Uptime: 99.97, LatencyMS: &latencies[1], LastCheck: "há 1 min", CreatedAt: now},
		{ID: "auth-service", Name: "Serviço de autenticação", URL: "https://auth.goping.dev/health", IntervalSeconds: 300, TimeoutSeconds: 10, Enabled: true, Status: "down", Uptime: 98.42, LastCheck: "há 2 min", CreatedAt: now},
		{ID: "docs", Name: "Documentação", URL: "https://docs.goping.dev", IntervalSeconds: 300, TimeoutSeconds: 10, Enabled: true, Status: "up", Uptime: 100, LatencyMS: &latencies[2], LastCheck: "há 3 min", CreatedAt: now},
	}
	activities := []Activity{
		{ID: "incident-auth", Type: "down", Title: "Serviço de autenticação está indisponível", Description: "A verificação excedeu o timeout de 10 segundos.", Time: "há 2 min"},
		{ID: "recovered-payments", Type: "up", Title: "API de pagamentos foi recuperada", Description: "O serviço voltou a responder com HTTP 200.", Time: "há 38 min"},
		{ID: "latency-website", Type: "info", Title: "Pico de latência detectado", Description: "Site institucional respondeu em 612 ms.", Time: "há 2 h"},
	}
	return monitors, activities
}
