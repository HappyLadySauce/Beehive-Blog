package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestNewAPICommandReturnsValidationError(t *testing.T) {
	t.Helper()

	viper.Reset()
	t.Cleanup(viper.Reset)

	configPath := filepath.Join(t.TempDir(), "beehive-blog.yaml")
	if err := os.WriteFile(configPath, []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write test config failed: %v", err)
	}

	cmd := NewAPICommand(context.Background(), "beehive-blog")
	cmd.SetArgs([]string{"--config", configPath, "--bind-address=", "--bind-port=0"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !strings.Contains(err.Error(), "bind-address is required") {
		t.Fatalf("expected bind-address validation error, got %v", err)
	}
	if !strings.Contains(err.Error(), "bind-port is required") {
		t.Fatalf("expected bind-port validation error, got %v", err)
	}
}
