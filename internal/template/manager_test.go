package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManager_List(t *testing.T) {
	// Create temp directory with test templates
	tmpDir, err := os.MkdirTemp("", "template-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test template files
	templates := []string{"default.typ", "modern.typ", "minimal.typ"}
	for _, name := range templates {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("// test"), 0644); err != nil {
			t.Fatalf("failed to create template file: %v", err)
		}
	}

	// Also create a non-.typ file that should be ignored
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("ignored"), 0644); err != nil {
		t.Fatalf("failed to create txt file: %v", err)
	}

	mgr := NewManager(tmpDir)
	list, err := mgr.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("expected 3 templates, got %d", len(list))
	}
}

func TestManager_List_EmptyDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "template-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	mgr := NewManager(tmpDir)
	list, err := mgr.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(list) != 0 {
		t.Errorf("expected 0 templates, got %d", len(list))
	}
}

func TestManager_List_NonExistentDir(t *testing.T) {
	mgr := NewManager("/nonexistent/path")
	list, err := mgr.List()
	if err != nil {
		t.Fatalf("List() should not fail for nonexistent dir: %v", err)
	}

	if list != nil {
		t.Errorf("expected nil list, got %v", list)
	}
}

func TestManager_Exists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "template-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a test template
	if err := os.WriteFile(filepath.Join(tmpDir, "default.typ"), []byte("// test"), 0644); err != nil {
		t.Fatalf("failed to create template file: %v", err)
	}

	mgr := NewManager(tmpDir)

	if !mgr.Exists("default") {
		t.Error("Exists() should return true for 'default'")
	}

	if mgr.Exists("nonexistent") {
		t.Error("Exists() should return false for 'nonexistent'")
	}
}

func TestManager_GetPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "template-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a test template
	expectedPath := filepath.Join(tmpDir, "default.typ")
	if err := os.WriteFile(expectedPath, []byte("// test"), 0644); err != nil {
		t.Fatalf("failed to create template file: %v", err)
	}

	mgr := NewManager(tmpDir)

	path := mgr.GetPath("default")
	if path != expectedPath {
		t.Errorf("GetPath() = %s, want %s", path, expectedPath)
	}

	path = mgr.GetPath("nonexistent")
	if path != "" {
		t.Errorf("GetPath() should return empty for nonexistent, got %s", path)
	}
}

func TestDefaultTemplateContent(t *testing.T) {
	content := DefaultTemplateContent()

	if len(content) == 0 {
		t.Error("DefaultTemplateContent() should return non-empty content")
	}

	// Check for some expected content
	if !contains(content, "sys.inputs.data") {
		t.Error("DefaultTemplateContent() should contain sys.inputs.data")
	}

	if !contains(content, "INVOICE") {
		t.Error("DefaultTemplateContent() should contain INVOICE")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
