// Copyright 2025 Zaytra
// SPDX-License-Identifier: Apache-2.0

package qs

// Format represents the encoding format for query strings
type Format string

const (
	FormatRFC1738 Format = "RFC1738"
	FormatRFC3986 Format = "RFC3986"
)

const DefaultFormat = FormatRFC3986
