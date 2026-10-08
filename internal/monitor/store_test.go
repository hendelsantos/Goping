package monitor

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestStorePersistsCreatedMonitor(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "goping.json")
	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	created, err := store.Create(CreateInput{
		Name:            "API pública",
		URL:             "https://example.com/health",
		IntervalSeconds: 300,
		TimeoutSeconds:  10,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Status != "unknown" {
		t.Fatalf("Create() status = %q, want unknown", created.Status)
	}

	reloaded, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore() after persistence error = %v", err)
	}
	monitors := reloaded.List()
	if len(monitors) != 5 {
		t.Fatalf("List() length = %d, want 5", len(monitors))
	}
	if monitors[0].ID != created.ID {
		t.Fatalf("List()[0].ID = %q, want %q", monitors[0].ID, created.ID)
	}
}

func TestStoreRejectsInvalidMonitor(t *testing.T) {
	t.Parallel()

	store, err := NewStore(filepath.Join(t.TempDir(), "goping.json"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	tests := []struct {
		name  string
		input CreateInput
	}{
		{name: "empty name", input: CreateInput{URL: "https://example.com", IntervalSeconds: 300, TimeoutSeconds: 10}},
		{name: "invalid scheme", input: CreateInput{Name: "FTP", URL: "ftp://example.com", IntervalSeconds: 300, TimeoutSeconds: 10}},
		{name: "embedded credentials", input: CreateInput{Name: "Secret", URL: "https://user:pass@example.com", IntervalSeconds: 300, TimeoutSeconds: 10}},
		{name: "short interval", input: CreateInput{Name: "Fast", URL: "https://example.com", IntervalSeconds: 60, TimeoutSeconds: 10}},
		{name: "long timeout", input: CreateInput{Name: "Slow", URL: "https://example.com", IntervalSeconds: 300, TimeoutSeconds: 60}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := store.Create(test.input); err == nil {
				t.Fatal("Create() error = nil, want validation error")
			}
		})
	}
}

func TestStoreDeleteMissingMonitor(t *testing.T) {
	t.Parallel()

	store, err := NewStore(filepath.Join(t.TempDir(), "goping.json"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	if err := store.Delete("missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Delete() error = %v, want ErrNotFound", err)
	}
}
