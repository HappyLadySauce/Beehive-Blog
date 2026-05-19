package driver

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestDriverRegistryCreateBackendUnsupported(t *testing.T) {
	reg := NewDriverRegistry()
	_, err := reg.CreateBackend("unknown", json.RawMessage(`{}`))
	if !errors.Is(err, ErrUnsupportedDriver) {
		t.Fatalf("CreateBackend error = %v, want ErrUnsupportedDriver", err)
	}
}

func TestDriverRegistryCreateBackendNilRegistry(t *testing.T) {
	var reg *DriverRegistry
	_, err := reg.CreateBackend(DriverLocal, json.RawMessage(`{"root":"."}`))
	if err == nil {
		t.Fatal("CreateBackend(nil registry) = nil, want error")
	}
}

func TestDriverRegistryRegisterAndNames(t *testing.T) {
	reg := NewDriverRegistry()
	reg.Register(DriverLocal, NewLocalDriver)
	names := reg.Names()
	if len(names) != 1 {
		t.Fatalf("Names() len = %d, want 1", len(names))
	}
	if names[0] != DriverLocal {
		t.Fatalf("Names() = %v, want single entry %s", names, DriverLocal)
	}
}

func TestNewLocalDriverRejectsEmptyRoot(t *testing.T) {
	_, err := NewLocalDriver(json.RawMessage(`{"root":""}`))
	if err == nil {
		t.Fatal("NewLocalDriver empty root = nil, want error")
	}
}

func TestNewLocalDriverRejectsInvalidJSON(t *testing.T) {
	_, err := NewLocalDriver(json.RawMessage(`not-json`))
	if err == nil {
		t.Fatal("NewLocalDriver invalid json = nil, want error")
	}
}
