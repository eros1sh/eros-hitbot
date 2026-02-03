// Package utils provides common utility functions used across the application.
// CODE QUALITY FIX: Centralized escape functions to avoid code duplication.
package utils

import "strings"

// EscapeJS escapes a string for safe use in JavaScript code.
// This function handles backslashes, single quotes, newlines, and carriage returns.
// Use this function when embedding Go strings into JavaScript code to prevent
// XSS attacks and syntax errors.
//
// Example:
//
//	script := fmt.Sprintf("var name = '%s';", utils.EscapeJS(userInput))
func EscapeJS(s string) string {
	return strings.NewReplacer(
		"\\", "\\\\",
		"'", "\\'",
		"\"", "\\\"",
		"\n", "\\n",
		"\r", "",
		"\t", "\\t",
		"\u2028", "\\u2028", // Line separator
		"\u2029", "\\u2029", // Paragraph separator
	).Replace(s)
}

// EscapeJSSingleQuote escapes a string for use in single-quoted JavaScript strings.
// This is a lighter version that only escapes single quotes and backslashes.
func EscapeJSSingleQuote(s string) string {
	return strings.NewReplacer(
		"\\", "\\\\",
		"'", "\\'",
		"\n", "\\n",
		"\r", "",
	).Replace(s)
}

// EscapeHTML escapes a string for safe use in HTML content.
// This prevents XSS attacks when embedding user input in HTML.
func EscapeHTML(s string) string {
	return strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&#39;",
	).Replace(s)
}

// SanitizePath removes potentially dangerous path components.
// Use this when handling user-provided file paths.
func SanitizePath(s string) string {
	// Remove null bytes
	s = strings.ReplaceAll(s, "\x00", "")
	// Remove path traversal attempts
	s = strings.ReplaceAll(s, "..", "")
	// Remove leading slashes
	s = strings.TrimLeft(s, "/\\")
	return s
}
