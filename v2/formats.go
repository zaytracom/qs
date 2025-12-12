// Copyright 2025 Zaytra
// SPDX-License-Identifier: Apache-2.0

package qs

import "strings"

// Format represents the encoding format for query strings.
// RFC1738 encodes spaces as '+', while RFC3986 encodes spaces as '%20'.
type Format string

const (
	// FormatRFC1738 encodes spaces as '+' (application/x-www-form-urlencoded)
	FormatRFC1738 Format = "RFC1738"
	// FormatRFC3986 encodes spaces as '%20' (standard percent-encoding)
	FormatRFC3986 Format = "RFC3986"
)

// DefaultFormat is the default encoding format (RFC3986)
const DefaultFormat = FormatRFC3986

// FormatterFunc is a function that formats an encoded string according to a specific RFC.
type FormatterFunc func(string) string

// Formatters maps Format constants to their formatter functions.
// RFC1738 converts %20 to +, RFC3986 returns the value unchanged.
var Formatters = map[Format]FormatterFunc{
	FormatRFC1738: formatRFC1738,
	FormatRFC3986: formatRFC3986,
}

// formatRFC1738 converts percent-encoded spaces (%20) to plus signs (+)
// as specified by RFC 1738 / application/x-www-form-urlencoded.
func formatRFC1738(value string) string {
	return strings.ReplaceAll(value, "%20", "+")
}

// formatRFC3986 returns the value unchanged.
// RFC 3986 uses %20 for spaces, which is the standard percent-encoding output.
func formatRFC3986(value string) string {
	return value
}

// GetFormatter returns the formatter function for the given format.
// Returns the RFC3986 formatter if the format is unknown.
func GetFormatter(format Format) FormatterFunc {
	if f, ok := Formatters[format]; ok {
		return f
	}
	return Formatters[DefaultFormat]
}
