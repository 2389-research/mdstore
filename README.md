# mdstore

A pure-functional Go library for file-based storage with markdown frontmatter, atomic writes, and cross-platform locking.

## Install

```bash
go get github.com/harperreed/mdstore
```

## API

### Atomic File Operations

```go
// Write data safely via temp file + rename. Creates parent dirs automatically.
mdstore.AtomicWrite("data/notes/hello.md", []byte("# Hello"))

// Ensure a directory exists (mkdir -p).
mdstore.EnsureDir("data/notes")

// Exclusive file-based lock (uses flock on Unix, retry loop on Windows).
mdstore.WithLock("data/", func() error {
    // critical section
    return nil
})
```

### YAML

```go
// Read YAML into a struct. Returns nil (not an error) if the file doesn't exist.
var cfg Config
mdstore.ReadYAML("config.yaml", &cfg)

// Write a struct as YAML atomically.
mdstore.WriteYAML("config.yaml", cfg)

// Append an item to a YAML list file.
mdstore.AppendYAML("log.yaml", entry)
```

### Markdown Frontmatter

```go
// Split frontmatter from body.
yaml, body := mdstore.ParseFrontmatter("---\ntitle: Hello\n---\n# Content")

// Render metadata + body into a frontmatter document.
out, err := mdstore.RenderFrontmatter(meta, "# Content")
```

### Slugs

```go
mdstore.Slugify("Hello, World!")        // "hello-world"
mdstore.Slugify("")                     // "untitled"

// Collision-safe slugs.
mdstore.UniqueSlug("hello", exists)     // "hello-2" if "hello" is taken
```

### Time

```go
mdstore.FormatTime(time.Now())          // RFC3339Nano string
mdstore.ParseTime("2024-01-01T00:00:00Z") // flexible (RFC3339 or RFC3339Nano)
```

## Design

- **No state** -- every function is standalone, no structs or interfaces to wire up.
- **Atomic writes** -- temp file, fsync, rename. No partial writes.
- **Cross-platform locking** -- `syscall.Flock` on Unix, `O_CREATE|O_EXCL` retry loop on Windows.
- **Idempotent reads** -- `ReadYAML` returns nil for missing files instead of erroring.
- **Frontmatter normalization** -- handles `\r\n` line endings automatically.

## Dependencies

- [gopkg.in/yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3) for YAML marshaling
- Go stdlib for everything else

## License

See [LICENSE](LICENSE) for details.
