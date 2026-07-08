package hot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHotCapabilityDoesNotReturnToGlobalService(t *testing.T) {
	serviceDir := filepath.Clean("../service")
	entries, err := os.ReadDir(serviceDir)
	if err != nil {
		t.Fatalf("read service dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		if entry.Name() == "setting_service.go" || entry.Name() == "setting_shell_config.go" {
			continue
		}
		name := strings.ToLower(entry.Name())
		if strings.Contains(name, "hot_content") || strings.Contains(name, "hot_topic") || name == "hot.go" {
			t.Fatalf("hot capability file must stay under internal/hot, found internal/service/%s", entry.Name())
		}

		raw, err := os.ReadFile(filepath.Join(serviceDir, entry.Name()))
		if err != nil {
			t.Fatalf("read service file %s: %v", entry.Name(), err)
		}
		body := string(raw)
		for _, marker := range []string{
			"HotContent",
			"HotTopic",
			"hot_content",
			"hot_topics",
			"HOT_CONTENT_",
		} {
			if strings.Contains(body, marker) {
				t.Fatalf("hot capability marker %q must stay under internal/hot, found internal/service/%s", marker, entry.Name())
			}
		}
	}
}
