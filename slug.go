// ABOUTME: Text utilities for generating URL-safe slugs.
// ABOUTME: Provides Slugify for converting strings and UniqueSlug for collision avoidance.
package mdstore

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	nonAlphanumeric    = regexp.MustCompile(`[^a-z0-9]+`)
	consecutiveHyphens = regexp.MustCompile(`-{2,}`)
)

// Slugify converts a string to a URL-safe slug.
// Lowercases, replaces non-alphanumeric with hyphens, collapses consecutive hyphens,
// trims leading/trailing hyphens. Returns "untitled" if result is empty.
func Slugify(s string) string {
	slug := strings.ToLower(s)
	slug = nonAlphanumeric.ReplaceAllString(slug, "-")
	slug = consecutiveHyphens.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	if slug == "" {
		return "untitled"
	}

	return slug
}

// UniqueSlug returns a slug that doesn't collide. Calls exists(candidate) to check.
// If the base slug collides, appends "-2", "-3", etc. until unique.
func UniqueSlug(s string, exists func(string) bool) string {
	base := Slugify(s)

	if !exists(base) {
		return base
	}

	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if !exists(candidate) {
			return candidate
		}
	}
}
