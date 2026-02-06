# mdstore: Shared Markdown Storage Utilities

Thin utility library for file-based storage backends across the suite. Pure functions, no framework.

## API (12 functions)

### File Operations
- `AtomicWrite(path string, data []byte) error` — write to tmp + rename
- `WithLock(dir string, fn func() error) error` — flock serialization
- `EnsureDir(path string) error` — mkdir -p

### YAML Helpers
- `ReadYAML(path string, dest interface{}) error` — read + unmarshal
- `WriteYAML(path string, src interface{}) error` — marshal + atomic write
- `AppendYAML(path string, item interface{}) error` — read slice, append, write

### Markdown Frontmatter
- `ParseFrontmatter(content string) (metadata map[string]string, body string, err error)`
- `RenderFrontmatter(metadata interface{}, body string) (string, error)`

### Text Utilities
- `Slugify(s string) string` — lowercase, alphanumeric + hyphens
- `UniqueSlug(s string, exists func(string) bool) string` — collision-safe

### Time
- `FormatTime(t time.Time) string` — RFC3339Nano
- `ParseTime(s string) (time.Time, error)` — flexible RFC3339/RFC3339Nano

## Dependencies

- `gopkg.in/yaml.v3`
- stdlib only otherwise

## Design Principles

- Pure functions, no structs, no state
- Platform-aware locking (Unix flock, Windows fallback)
- Every function is independently useful
- No opinions about data models or file layouts
