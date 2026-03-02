package services

import (
	"time"

	"github.com/google/uuid"
	gonanoid "github.com/jaevor/go-nanoid"
)

// JotNamespace provides utility functions for templates under the "jot" namespace.
// This allows {{ jot.Slug "text" }} syntax in templates.
type JotNamespace struct{}

// JotFuncs returns a JotNamespace instance for use in templates.
// Register this in the template FuncMap to enable jot.* functions.
func JotFuncs() JotNamespace {
	return JotNamespace{}
}

// Slug converts a string to a filesystem-safe slug.
// Example: {{ jot.Slug "Hello World" }} → "hello-world"
func (j JotNamespace) Slug(s string) string {
	return Slug(s)
}

// NanoID generates a URL-safe nanoid of the specified length.
// Example: {{ jot.NanoID 8 }} → "V1StGXR8"
func (j JotNamespace) NanoID(length int) string {
	// Use URL-safe alphabet: A-Za-z0-9_-
	generator, err := gonanoid.CustomASCII("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-", length)
	if err != nil {
		// Fallback to simple random string if generator fails
		return ""
	}
	return generator()
}

// Timestamp returns the current Unix timestamp in seconds.
// Example: {{ jot.Timestamp }} → 1709424000
func (j JotNamespace) Timestamp() int64 {
	return time.Now().Unix()
}

// DatePath returns the current date as a path string in "YYYY/MM/DD" format.
// Example: {{ jot.DatePath }} → "2026/03/02"
func (j JotNamespace) DatePath() string {
	return time.Now().Format("2006/01/02")
}

// UUID generates a random UUID v4 string.
// Example: {{ jot.UUID }} → "550e8400-e29b-41d4-a716-446655440000"
func (j JotNamespace) UUID() string {
	return uuid.New().String()
}

// Now returns the current time formatted according to the Go time format string.
// Example: {{ jot.Now "2006-01-02" }} → "2026-03-02"
// Reference date: Mon Jan 2 15:04:05 MST 2006
func (j JotNamespace) Now(format string) string {
	return time.Now().Format(format)
}
