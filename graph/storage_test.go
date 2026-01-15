package graph

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJsonStorageSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	g := NewGraph()
	v1, _ := g.AddVertex("https://example.com/a")
	v2, _ := g.AddVertex("https://example.com/b")
	g.AddEdge(v1, v2)
	storage := &JsonStorage{filename: path}
	if err := storage.Save(g); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	loaded, err := storage.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(loaded.Vertices) != 2 {
		t.Errorf("expected 2 vertices, got %d", len(loaded.Vertices))
	}
}

func TestJsonStorageLoadPreservesExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	g := NewGraph()
	g.AddVertex("https://example.com")
	storage := &JsonStorage{filename: path}
	storage.Save(g)
	info, _ := os.Stat(path)
	sizeBefore := info.Size()
	storage.Load()
	info, _ = os.Stat(path)
	sizeAfter := info.Size()
	if sizeAfter != sizeBefore {
		t.Errorf("Load changed file size from %d to %d", sizeBefore, sizeAfter)
	}
}

func TestBinaryStorageSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.gob")
	g := NewGraph()
	g.AddVertex("https://example.com/a")
	storage := &BinaryStorage{filename: path}
	if err := storage.Save(g); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	loaded, err := storage.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(loaded.Vertices) != 1 {
		t.Errorf("expected 1 vertex, got %d", len(loaded.Vertices))
	}
}

func TestJsonStorageLoadNonexistentFile(t *testing.T) {
	storage := &JsonStorage{filename: "/nonexistent/path.json"}
	_, err := storage.Load()
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestNewStorageJson(t *testing.T) {
	s := NewStorage("test.json", "json")
	if _, ok := s.(*JsonStorage); !ok {
		t.Error("expected JsonStorage for json format")
	}
}

func TestNewStorageBinary(t *testing.T) {
	s := NewStorage("test.gob", "binary")
	if _, ok := s.(*BinaryStorage); !ok {
		t.Error("expected BinaryStorage for binary format")
	}
}
