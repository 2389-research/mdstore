// ABOUTME: Comprehensive tests for all mdstore functions.
// ABOUTME: Covers atomic writes, locking, YAML ops, frontmatter, slugs, and time helpers.
package mdstore

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// --- AtomicWrite tests ---

func TestAtomicWrite_Basic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	err := AtomicWrite(path, []byte("hello world"))
	if err != nil {
		t.Fatalf("AtomicWrite failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(data) != "hello world" {
		t.Errorf("got %q, want %q", string(data), "hello world")
	}
}

func TestAtomicWrite_Overwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	if err := AtomicWrite(path, []byte("first")); err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	if err := AtomicWrite(path, []byte("second")); err != nil {
		t.Fatalf("second write failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(data) != "second" {
		t.Errorf("got %q, want %q", string(data), "second")
	}
}

func TestAtomicWrite_CreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "c", "test.txt")

	err := AtomicWrite(path, []byte("nested"))
	if err != nil {
		t.Fatalf("AtomicWrite with nested dirs failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(data) != "nested" {
		t.Errorf("got %q, want %q", string(data), "nested")
	}
}

// --- EnsureDir tests ---

func TestEnsureDir_Create(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "newdir")

	if err := EnsureDir(path); err != nil {
		t.Fatalf("EnsureDir failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestEnsureDir_Idempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "newdir")

	if err := EnsureDir(path); err != nil {
		t.Fatalf("first EnsureDir failed: %v", err)
	}

	if err := EnsureDir(path); err != nil {
		t.Fatalf("second EnsureDir failed: %v", err)
	}
}

func TestEnsureDir_Nested(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "c")

	if err := EnsureDir(path); err != nil {
		t.Fatalf("EnsureDir nested failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if !info.IsDir() {
		t.Error("expected directory")
	}
}

// --- WithLock tests ---

func TestWithLock_Basic(t *testing.T) {
	dir := t.TempDir()
	called := false

	err := WithLock(dir, func() error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("WithLock failed: %v", err)
	}

	if !called {
		t.Error("function was not called")
	}
}

func TestWithLock_ConcurrentGoroutines(t *testing.T) {
	dir := t.TempDir()
	var counter int64
	var maxConcurrent int64
	var currentConcurrent int64
	var wg sync.WaitGroup

	const goroutines = 10

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := WithLock(dir, func() error {
				cur := atomic.AddInt64(&currentConcurrent, 1)
				// Track maximum concurrency observed
				for {
					old := atomic.LoadInt64(&maxConcurrent)
					if cur <= old || atomic.CompareAndSwapInt64(&maxConcurrent, old, cur) {
						break
					}
				}
				// Simulate some work
				time.Sleep(5 * time.Millisecond)
				atomic.AddInt64(&counter, 1)
				atomic.AddInt64(&currentConcurrent, -1)
				return nil
			})
			if err != nil {
				t.Errorf("WithLock failed: %v", err)
			}
		}()
	}

	wg.Wait()

	if atomic.LoadInt64(&counter) != goroutines {
		t.Errorf("expected counter=%d, got %d", goroutines, atomic.LoadInt64(&counter))
	}

	if atomic.LoadInt64(&maxConcurrent) > 1 {
		t.Errorf("lock did not serialize: max concurrent = %d", atomic.LoadInt64(&maxConcurrent))
	}
}

// --- ReadYAML tests ---

func TestReadYAML_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")

	content := "name: Alice\nage: 30\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	var result map[string]interface{}
	if err := ReadYAML(path, &result); err != nil {
		t.Fatalf("ReadYAML failed: %v", err)
	}

	if result["name"] != "Alice" {
		t.Errorf("got name=%v, want Alice", result["name"])
	}
	if result["age"] != 30 {
		t.Errorf("got age=%v, want 30", result["age"])
	}
}

func TestReadYAML_MissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.yaml")

	var result map[string]interface{}
	err := ReadYAML(path, &result)
	if err != nil {
		t.Fatalf("ReadYAML should return nil for missing file, got: %v", err)
	}

	if result != nil {
		t.Errorf("expected nil result for missing file, got %v", result)
	}
}

func TestReadYAML_Malformed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")

	if err := os.WriteFile(path, []byte(":::not valid yaml[[["), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	var result map[string]interface{}
	err := ReadYAML(path, &result)
	if err == nil {
		t.Error("expected error for malformed YAML, got nil")
	}
}

// --- WriteYAML tests ---

func TestWriteYAML_Basic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.yaml")

	data := map[string]string{"greeting": "hello"}
	if err := WriteYAML(path, data); err != nil {
		t.Fatalf("WriteYAML failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	var result map[string]string
	if err := yaml.Unmarshal(content, &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result["greeting"] != "hello" {
		t.Errorf("got %q, want %q", result["greeting"], "hello")
	}
}

func TestWriteYAML_Overwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.yaml")

	if err := WriteYAML(path, map[string]string{"k": "v1"}); err != nil {
		t.Fatalf("first WriteYAML failed: %v", err)
	}

	if err := WriteYAML(path, map[string]string{"k": "v2"}); err != nil {
		t.Fatalf("second WriteYAML failed: %v", err)
	}

	var result map[string]string
	if err := ReadYAML(path, &result); err != nil {
		t.Fatalf("ReadYAML failed: %v", err)
	}

	if result["k"] != "v2" {
		t.Errorf("got %q, want %q", result["k"], "v2")
	}
}

// --- AppendYAML tests ---

type testItem struct {
	Name  string `yaml:"name"`
	Value int    `yaml:"value"`
}

func TestAppendYAML_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "items.yaml")

	err := AppendYAML(path, testItem{Name: "first", Value: 1})
	if err != nil {
		t.Fatalf("AppendYAML failed: %v", err)
	}

	var items []testItem
	if err := ReadYAML(path, &items); err != nil {
		t.Fatalf("ReadYAML failed: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Name != "first" || items[0].Value != 1 {
		t.Errorf("unexpected item: %+v", items[0])
	}
}

func TestAppendYAML_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "items.yaml")

	if err := AppendYAML(path, testItem{Name: "first", Value: 1}); err != nil {
		t.Fatalf("first AppendYAML failed: %v", err)
	}

	if err := AppendYAML(path, testItem{Name: "second", Value: 2}); err != nil {
		t.Fatalf("second AppendYAML failed: %v", err)
	}

	var items []testItem
	if err := ReadYAML(path, &items); err != nil {
		t.Fatalf("ReadYAML failed: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[1].Name != "second" || items[1].Value != 2 {
		t.Errorf("unexpected second item: %+v", items[1])
	}
}

func TestAppendYAML_Multiple(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "items.yaml")

	for i := 0; i < 5; i++ {
		if err := AppendYAML(path, testItem{Name: "item", Value: i}); err != nil {
			t.Fatalf("AppendYAML iteration %d failed: %v", i, err)
		}
	}

	var items []testItem
	if err := ReadYAML(path, &items); err != nil {
		t.Fatalf("ReadYAML failed: %v", err)
	}

	if len(items) != 5 {
		t.Fatalf("expected 5 items, got %d", len(items))
	}

	for i, item := range items {
		if item.Value != i {
			t.Errorf("item %d: got value=%d, want %d", i, item.Value, i)
		}
	}
}

// --- ParseFrontmatter tests ---

func TestParseFrontmatter_WithFrontmatter(t *testing.T) {
	content := "---\ntitle: Hello\ntags: [a, b]\n---\nThis is the body."

	yamlStr, body := ParseFrontmatter(content)

	if yamlStr == "" {
		t.Fatal("expected non-empty yaml")
	}

	var meta map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &meta); err != nil {
		t.Fatalf("failed to unmarshal yaml: %v", err)
	}

	if meta["title"] != "Hello" {
		t.Errorf("got title=%v, want Hello", meta["title"])
	}

	if body != "This is the body." {
		t.Errorf("got body=%q, want %q", body, "This is the body.")
	}
}

func TestParseFrontmatter_WithoutFrontmatter(t *testing.T) {
	content := "Just regular markdown content."

	yamlStr, body := ParseFrontmatter(content)

	if yamlStr != "" {
		t.Errorf("expected empty yaml, got %q", yamlStr)
	}

	if body != content {
		t.Errorf("got body=%q, want %q", body, content)
	}
}

func TestParseFrontmatter_Empty(t *testing.T) {
	yamlStr, body := ParseFrontmatter("")

	if yamlStr != "" {
		t.Errorf("expected empty yaml, got %q", yamlStr)
	}

	if body != "" {
		t.Errorf("expected empty body, got %q", body)
	}
}

func TestParseFrontmatter_NoClosingDelimiter(t *testing.T) {
	content := "---\ntitle: Hello\nno closing delimiter here"

	yamlStr, body := ParseFrontmatter(content)

	if yamlStr != "" {
		t.Errorf("expected empty yaml for malformed frontmatter, got %q", yamlStr)
	}

	if body != content {
		t.Errorf("expected full content as body, got %q", body)
	}
}

func TestParseFrontmatter_MultilineBody(t *testing.T) {
	content := "---\ntitle: Test\n---\nLine 1\nLine 2\nLine 3"

	yamlStr, body := ParseFrontmatter(content)

	if yamlStr == "" {
		t.Fatal("expected non-empty yaml")
	}

	expected := "Line 1\nLine 2\nLine 3"
	if body != expected {
		t.Errorf("got body=%q, want %q", body, expected)
	}
}

// --- RenderFrontmatter tests ---

func TestRenderFrontmatter_Basic(t *testing.T) {
	meta := map[string]string{"title": "Hello"}
	result, err := RenderFrontmatter(meta, "Body text\n")
	if err != nil {
		t.Fatalf("RenderFrontmatter failed: %v", err)
	}

	if !strings.HasPrefix(result, "---\n") {
		t.Error("expected result to start with ---")
	}

	if !strings.Contains(result, "title: Hello") {
		t.Error("expected result to contain 'title: Hello'")
	}

	if !strings.HasSuffix(result, "Body text\n") {
		t.Errorf("expected result to end with body, got %q", result)
	}
}

func TestRenderFrontmatter_ComplexStruct(t *testing.T) {
	type meta struct {
		Title string   `yaml:"title"`
		Tags  []string `yaml:"tags"`
		Count int      `yaml:"count"`
	}

	m := meta{
		Title: "Complex",
		Tags:  []string{"a", "b", "c"},
		Count: 42,
	}

	result, err := RenderFrontmatter(m, "Some body.\n")
	if err != nil {
		t.Fatalf("RenderFrontmatter failed: %v", err)
	}

	if !strings.Contains(result, "title: Complex") {
		t.Error("expected title in output")
	}
	if !strings.Contains(result, "count: 42") {
		t.Error("expected count in output")
	}
}

// --- Slugify tests ---

func TestSlugify_Basic(t *testing.T) {
	result := Slugify("Hello World")
	if result != "hello-world" {
		t.Errorf("got %q, want %q", result, "hello-world")
	}
}

func TestSlugify_SpecialChars(t *testing.T) {
	result := Slugify("Hello, World! How's it going?")
	if result != "hello-world-how-s-it-going" {
		t.Errorf("got %q, want %q", result, "hello-world-how-s-it-going")
	}
}

func TestSlugify_Unicode(t *testing.T) {
	result := Slugify("Caf\u00e9 au lait")
	// Unicode non-ASCII chars become hyphens
	if result != "caf-au-lait" {
		t.Errorf("got %q, want %q", result, "caf-au-lait")
	}
}

func TestSlugify_Empty(t *testing.T) {
	result := Slugify("")
	if result != "untitled" {
		t.Errorf("got %q, want %q", result, "untitled")
	}
}

func TestSlugify_OnlySpecialChars(t *testing.T) {
	result := Slugify("@#$%^&*()")
	if result != "untitled" {
		t.Errorf("got %q, want %q", result, "untitled")
	}
}

func TestSlugify_AlreadyClean(t *testing.T) {
	result := Slugify("already-clean")
	if result != "already-clean" {
		t.Errorf("got %q, want %q", result, "already-clean")
	}
}

func TestSlugify_ConsecutiveHyphens(t *testing.T) {
	result := Slugify("hello---world")
	if result != "hello-world" {
		t.Errorf("got %q, want %q", result, "hello-world")
	}
}

// --- UniqueSlug tests ---

func TestUniqueSlug_NoCollision(t *testing.T) {
	result := UniqueSlug("Hello World", func(s string) bool {
		return false // nothing exists
	})

	if result != "hello-world" {
		t.Errorf("got %q, want %q", result, "hello-world")
	}
}

func TestUniqueSlug_WithCollision(t *testing.T) {
	existing := map[string]bool{
		"hello-world":   true,
		"hello-world-2": true,
	}

	result := UniqueSlug("Hello World", func(s string) bool {
		return existing[s]
	})

	if result != "hello-world-3" {
		t.Errorf("got %q, want %q", result, "hello-world-3")
	}
}

func TestUniqueSlug_FirstCollision(t *testing.T) {
	existing := map[string]bool{
		"test": true,
	}

	result := UniqueSlug("test", func(s string) bool {
		return existing[s]
	})

	if result != "test-2" {
		t.Errorf("got %q, want %q", result, "test-2")
	}
}

// --- FormatTime / ParseTime tests ---

func TestFormatTime_ParseTime_RoundTrip(t *testing.T) {
	now := time.Now().UTC()
	formatted := FormatTime(now)

	parsed, err := ParseTime(formatted)
	if err != nil {
		t.Fatalf("ParseTime failed: %v", err)
	}

	if !now.Equal(parsed) {
		t.Errorf("round-trip failed: got %v, want %v", parsed, now)
	}
}

func TestParseTime_RFC3339(t *testing.T) {
	input := "2024-01-15T10:30:00Z"

	parsed, err := ParseTime(input)
	if err != nil {
		t.Fatalf("ParseTime failed: %v", err)
	}

	if parsed.Year() != 2024 || parsed.Month() != 1 || parsed.Day() != 15 {
		t.Errorf("unexpected date: %v", parsed)
	}
}

func TestParseTime_RFC3339Nano(t *testing.T) {
	input := "2024-01-15T10:30:00.123456789Z"

	parsed, err := ParseTime(input)
	if err != nil {
		t.Fatalf("ParseTime failed: %v", err)
	}

	if parsed.Nanosecond() != 123456789 {
		t.Errorf("expected nanoseconds=123456789, got %d", parsed.Nanosecond())
	}
}

func TestParseTime_InvalidFormat(t *testing.T) {
	_, err := ParseTime("not-a-time")
	if err == nil {
		t.Error("expected error for invalid time format")
	}

	if !strings.Contains(err.Error(), "mdstore") {
		t.Errorf("error should mention mdstore: %v", err)
	}
}

func TestFormatTime_IsRFC3339Nano(t *testing.T) {
	now := time.Date(2024, 6, 15, 12, 0, 0, 123456789, time.UTC)
	formatted := FormatTime(now)

	if !strings.Contains(formatted, "123456789") {
		t.Errorf("FormatTime should include nanoseconds: %s", formatted)
	}
}
