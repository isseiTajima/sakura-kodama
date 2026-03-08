package internal_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	repoRoot             = ".."
	wailsConfigPath      = "wails.json"
	statusDocPath        = "docs/STATUS.md"
	testScopeDocPath     = "docs/test-scope.md"
	expectedDevServerURL = "http://127.0.0.1:5173"
)

type fakeDevWorkflow struct {
	serverURL string
}

func (wf *fakeDevWorkflow) Configure(serverURL string) error {
	if serverURL == "" {
		return errors.New("server URL is required")
	}
	wf.serverURL = serverURL
	return nil
}

func TestDevWorkflowConfiguresProxyURL(t *testing.T) {
	t.Parallel()

	serverURL := loadDevServerURLFromFile(t, repoPath(wailsConfigPath))

	wf := &fakeDevWorkflow{}
	if err := wf.Configure(serverURL); err != nil {
		t.Fatalf("configure workflow: %v", err)
	}
	if wf.serverURL != expectedDevServerURL {
		t.Fatalf("workflow serverURL = %s, want %s", wf.serverURL, expectedDevServerURL)
	}
}

func TestDevWorkflowRejectsMissingOrMismatchedURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		payload       string
		wantSubstring string
	}{
		{
			name:          "missing",
			payload:       `{}`,
			wantSubstring: "frontend:dev:serverUrl is empty",
		},
		{
			name:    "mismatch",
			payload: `{"frontend:dev:serverUrl":"http://localhost:9999"}`,
			wantSubstring: fmt.Sprintf(
				"frontend:dev:serverUrl = %s, want %s",
				"http://localhost:9999",
				expectedDevServerURL,
			),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := loadDevServerURL([]byte(tt.payload))
			if err == nil {
				t.Fatalf("expected error for %s payload", tt.name)
			}
			if tt.wantSubstring != "" && !strings.Contains(err.Error(), tt.wantSubstring) {
				t.Fatalf("error %q does not contain %q", err, tt.wantSubstring)
			}
		})
	}
}

func TestStatusDocDocumentsDevWorkflow(t *testing.T) {
	t.Parallel()

	data := readFile(t, repoPath(statusDocPath))
	content := string(data)

	tokens := []string{
		"npm run dev -- --host 127.0.0.1 --port 5173",
		"make dev",
		"http://localhost:34115",
		"AssetServer options invalid",
		"appendStatusLog",
		"docs/test-scope.md",
		"make test",
	}

	for _, token := range tokens {
		if !strings.Contains(content, token) {
			t.Fatalf("%s must mention %q to document the dev workflow handshake", statusDocPath, token)
		}
	}
}

func TestTestScopeDocumentsDevWorkflowRegression(t *testing.T) {
	t.Parallel()

	data := readFile(t, repoPath(testScopeDocPath))
	content := string(data)
	tokens := []string{
		"wails_dev_workflow_test.go",
		"frontend:dev:serverUrl",
		"npm run dev -- --host 127.0.0.1 --port 5173",
		"make dev",
	}

	for _, token := range tokens {
		if !strings.Contains(content, token) {
			t.Fatalf("%s must mention %q to capture the regression guard", testScopeDocPath, token)
		}
	}
}

func loadDevServerURLFromFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	url, err := loadDevServerURL(data)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return url
}

func loadDevServerURL(data []byte) (string, error) {
	var cfg struct {
		ServerURL string `json:"frontend:dev:serverUrl"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("unmarshal wails.json: %w", err)
	}
	if cfg.ServerURL == "" {
		return "", errors.New("frontend:dev:serverUrl is empty")
	}
	if cfg.ServerURL != expectedDevServerURL {
		return "", fmt.Errorf("frontend:dev:serverUrl = %s, want %s", cfg.ServerURL, expectedDevServerURL)
	}
	return cfg.ServerURL, nil
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func repoPath(rel string) string {
	return filepath.Clean(filepath.Join(repoRoot, rel))
}
