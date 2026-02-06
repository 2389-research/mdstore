// ABOUTME: YAML file helpers for reading, writing, and appending YAML documents.
// ABOUTME: Uses AtomicWrite for safe writes and gopkg.in/yaml.v3 for marshaling.
package mdstore

import (
	"errors"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

// ReadYAML reads a YAML file and unmarshals into dest.
// Returns nil (not error) if the file doesn't exist.
func ReadYAML(path string, dest interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}

	return yaml.Unmarshal(data, dest)
}

// WriteYAML marshals src to YAML and writes atomically.
func WriteYAML(path string, src interface{}) error {
	data, err := yaml.Marshal(src)
	if err != nil {
		return err
	}

	return AtomicWrite(path, data)
}

// AppendYAML reads a YAML file as a slice of T, appends item, and writes back atomically.
// If the file doesn't exist, creates it with just [item].
func AppendYAML[T any](path string, item T) error {
	var existing []T

	if err := ReadYAML(path, &existing); err != nil {
		return err
	}

	existing = append(existing, item)

	return WriteYAML(path, existing)
}
