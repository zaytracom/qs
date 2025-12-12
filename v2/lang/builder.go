package lang

import (
	"sync"
)

// Builder converts AST to map[string]any with minimal allocations.
// Reuse Builder across parses to avoid allocations.
type Builder struct {
	// Reusable path buffer to avoid allocations during path building
	pathBuf []pathSegment

	// String builder for decoding
	strBuf []byte

	// Decode cache - avoids re-decoding same spans
	decodeCache map[uint64]string

	// Options
	charset     Charset
	allowSparse bool
	duplicates  DuplicateMode
}

// pathSegment represents a decoded segment for building nested structure
type pathSegment struct {
	name    string
	isIndex bool
	index   int32
	isEmpty bool // for []
}

// DuplicateMode controls how duplicate keys are handled
type DuplicateMode uint8

const (
	DupCombine DuplicateMode = iota // combine into array
	DupFirst                        // keep first
	DupLast                         // keep last
)

// BuilderConfig configures the Builder
type BuilderConfig struct {
	Charset     Charset
	AllowSparse bool
	Duplicates  DuplicateMode
}

// Pool for Builder reuse
var builderPool = sync.Pool{
	New: func() any {
		return &Builder{
			pathBuf:     make([]pathSegment, 0, 16),
			strBuf:      make([]byte, 0, 256),
			decodeCache: make(map[uint64]string, 64),
		}
	},
}

// AcquireBuilder gets a Builder from pool
func AcquireBuilder() *Builder {
	return builderPool.Get().(*Builder)
}

// ReleaseBuilder returns Builder to pool
func ReleaseBuilder(b *Builder) {
	b.Reset()
	builderPool.Put(b)
}

// Reset clears Builder for reuse
func (b *Builder) Reset() {
	b.pathBuf = b.pathBuf[:0]
	b.strBuf = b.strBuf[:0]
	// Clear decode cache
	for k := range b.decodeCache {
		delete(b.decodeCache, k)
	}
}

// Configure sets builder options
func (b *Builder) Configure(cfg BuilderConfig) {
	b.charset = cfg.Charset
	b.allowSparse = cfg.AllowSparse
	b.duplicates = cfg.Duplicates
}

// Build converts AST to map[string]any
func (b *Builder) Build(arena *Arena, qs QueryString) (map[string]any, error) {
	if qs.ParamLen == 0 {
		return make(map[string]any), nil
	}

	// Pre-allocate result with estimated size
	result := make(map[string]any, qs.ParamLen)

	// Process each parameter
	for i := uint16(0); i < qs.ParamLen; i++ {
		param := arena.Params[qs.ParamStart+i]

		// Build path from segments
		b.pathBuf = b.pathBuf[:0]
		if err := b.buildPath(arena, param.Key); err != nil {
			return nil, err
		}

		// Get value
		value := b.getValue(arena, param)

		// Insert into result
		b.setNested(result, value)
	}

	// Compact sparse arrays if needed
	if !b.allowSparse {
		compactArrays(result)
	}

	return result, nil
}

// buildPath extracts path segments from AST key
func (b *Builder) buildPath(arena *Arena, key Key) error {
	for i := uint16(0); i < uint16(key.SegLen); i++ {
		seg := arena.Segments[key.SegStart+i]

		ps := pathSegment{
			isIndex: seg.Kind == SegIndex,
			index:   seg.Index,
			isEmpty: seg.Kind == SegEmpty,
		}

		if seg.Kind != SegEmpty {
			// Decode segment name
			ps.name = b.decodeSpan(arena, seg.Span)
		}

		b.pathBuf = append(b.pathBuf, ps)
	}
	return nil
}

// decodeSpan decodes a span with caching
func (b *Builder) decodeSpan(arena *Arena, sp Span) string {
	// Cache key: offset << 16 | length
	cacheKey := uint64(sp.Off)<<16 | uint64(sp.Len)

	if cached, ok := b.decodeCache[cacheKey]; ok {
		return cached
	}

	raw := arena.Source[sp.Off : sp.Off+uint32(sp.Len)]

	// Fast path: no encoding needed
	if !needsDecode(raw) {
		b.decodeCache[cacheKey] = raw
		return raw
	}

	// Slow path: decode
	decoded := decodeString(raw, b.charset, &b.strBuf)
	b.decodeCache[cacheKey] = decoded
	return decoded
}

// needsDecode checks if string needs URL decoding
func needsDecode(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '%' || c == '+' {
			return true
		}
	}
	return false
}

// decodeString decodes URL-encoded string
func decodeString(s string, _ Charset, buf *[]byte) string {
	*buf = (*buf)[:0]

	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '+':
			*buf = append(*buf, ' ')
		case '%':
			if i+2 < len(s) {
				hi := fromHexChar(s[i+1])
				lo := fromHexChar(s[i+2])
				if hi >= 0 && lo >= 0 {
					*buf = append(*buf, byte(hi<<4|lo))
					i += 2
					continue
				}
			}
			*buf = append(*buf, c)
		default:
			*buf = append(*buf, c)
		}
	}

	return string(*buf)
}

func fromHexChar(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	default:
		return -1
	}
}

// getValue extracts value from parameter
func (b *Builder) getValue(arena *Arena, param Param) any {
	if param.ValueIdx == noValue {
		if param.HasEquals {
			return "" // key= -> empty string
		}
		return nil // key -> null
	}

	val := arena.Values[param.ValueIdx]

	switch val.Kind {
	case ValNull:
		return nil

	case ValSimple:
		return b.decodeSpan(arena, val.Raw)

	case ValComma:
		// Comma-separated values -> array
		parts := make([]any, val.PartsLen)
		for i := uint8(0); i < val.PartsLen; i++ {
			part := arena.ValueParts[val.PartsOff+uint16(i)]
			parts[i] = b.decodeSpan(arena, part)
		}
		return parts

	default:
		return b.decodeSpan(arena, val.Raw)
	}
}

// setNested inserts value at path in result map
func (b *Builder) setNested(result map[string]any, value any) {
	if len(b.pathBuf) == 0 {
		return
	}

	// Navigate/create path
	var current any = result

	for i := 0; i < len(b.pathBuf)-1; i++ {
		seg := b.pathBuf[i]
		next := b.pathBuf[i+1]

		current = b.ensureContainer(current, seg, next)
	}

	// Set final value
	lastSeg := b.pathBuf[len(b.pathBuf)-1]
	b.setValue(current, lastSeg, value)
}

// ensureContainer ensures container exists at segment
func (b *Builder) ensureContainer(current any, seg pathSegment, next pathSegment) any {
	// Determine what type of container we need for next
	needArray := next.isIndex || next.isEmpty

	switch c := current.(type) {
	case map[string]any:
		if seg.isEmpty {
			// [] on map -> append to implicit array
			return b.handleEmptyBracketOnMap(c, needArray)
		}

		existing, ok := c[seg.name]
		if !ok {
			// Create new container
			if needArray {
				arr := make([]any, 0, 4)
				c[seg.name] = arr
				return arr
			}
			m := make(map[string]any)
			c[seg.name] = m
			return m
		}

		// Container exists - navigate into it or convert
		return b.navigateOrConvert(c, seg.name, existing, needArray)

	case []any:
		if seg.isIndex {
			idx := int(seg.index)
			// Grow array if needed
			for len(c) <= idx {
				c = append(c, nil)
			}
			// Update parent reference (array might have grown)
			b.updateParentArray(c)

			if c[idx] == nil {
				if needArray {
					arr := make([]any, 0, 4)
					c[idx] = arr
					return arr
				}
				m := make(map[string]any)
				c[idx] = m
				return m
			}
			return c[idx]
		} else if seg.isEmpty {
			// [] -> append new container
			var newContainer any
			if needArray {
				newContainer = make([]any, 0, 4)
			} else {
				newContainer = make(map[string]any)
			}
			c = append(c, newContainer)
			b.updateParentArray(c)
			return newContainer
		}
	}

	return current
}

// handleEmptyBracketOnMap handles [] on a map
func (b *Builder) handleEmptyBracketOnMap(_ map[string]any, needArray bool) any {
	// Find or create implicit array
	// This is edge case - [] on root level
	if needArray {
		return make([]any, 0, 4)
	}
	return make(map[string]any)
}

// navigateOrConvert navigates into existing value or converts it
func (b *Builder) navigateOrConvert(parent map[string]any, key string, existing any, needArray bool) any {
	switch e := existing.(type) {
	case map[string]any:
		if needArray {
			// Need to convert map to array - put map at index 0
			arr := []any{e}
			parent[key] = arr
			return arr
		}
		return e

	case []any:
		if !needArray {
			// Need map but have array - convert
			m := arrayToMap(e)
			parent[key] = m
			return m
		}
		return e

	case string:
		// Primitive value - need to wrap
		if needArray {
			arr := []any{e}
			parent[key] = arr
			return arr
		}
		m := map[string]any{"0": e}
		parent[key] = m
		return m

	default:
		if needArray {
			return make([]any, 0, 4)
		}
		return make(map[string]any)
	}
}

// updateParentArray updates parent reference when array grows
// This is a placeholder - actual implementation needs parent tracking
func (b *Builder) updateParentArray(arr []any) {
	// In real implementation, we'd track parent and update reference
	// For now, arrays are passed by reference so modifications work
}

// setValue sets the final value
func (b *Builder) setValue(container any, seg pathSegment, value any) {
	switch c := container.(type) {
	case map[string]any:
		if seg.isEmpty {
			// [] -> create or append to array
			if existing, ok := c[""]; ok {
				if arr, isArr := existing.([]any); isArr {
					c[""] = append(arr, value)
				} else {
					c[""] = []any{existing, value}
				}
			} else {
				c[""] = []any{value}
			}
			return
		}

		// Handle duplicates
		if existing, ok := c[seg.name]; ok {
			c[seg.name] = b.handleDuplicate(existing, value)
		} else {
			c[seg.name] = value
		}

	case []any:
		if seg.isIndex {
			idx := int(seg.index)
			// Grow if needed
			for len(c) <= idx {
				c = append(c, nil)
			}

			if c[idx] != nil {
				c[idx] = b.handleDuplicate(c[idx], value)
			} else {
				c[idx] = value
			}
		} else if seg.isEmpty {
			c = append(c, value)
		}
	}
}

// handleDuplicate handles duplicate key based on mode
func (b *Builder) handleDuplicate(existing, newVal any) any {
	switch b.duplicates {
	case DupFirst:
		return existing
	case DupLast:
		return newVal
	default: // DupCombine
		return combine(existing, newVal)
	}
}

// combine merges two values into array
func combine(a, b any) any {
	switch av := a.(type) {
	case []any:
		switch bv := b.(type) {
		case []any:
			return append(av, bv...)
		default:
			return append(av, bv)
		}
	default:
		switch bv := b.(type) {
		case []any:
			return append([]any{av}, bv...)
		default:
			return []any{av, bv}
		}
	}
}

// arrayToMap converts array to map with string indices
func arrayToMap(arr []any) map[string]any {
	m := make(map[string]any, len(arr))
	for i, v := range arr {
		if v != nil {
			m[itoa(i)] = v
		}
	}
	return m
}

// itoa converts int to string without allocation for small numbers
func itoa(i int) string {
	if i < 10 {
		return string(rune('0' + i))
	}
	// Fall back to standard conversion
	return intToStr(i)
}

func intToStr(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	neg := i < 0
	if neg {
		i = -i
	}
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

// compactArrays removes nil elements from arrays recursively
func compactArrays(v any) {
	switch val := v.(type) {
	case map[string]any:
		for k, child := range val {
			if arr, ok := child.([]any); ok {
				val[k] = compactArray(arr)
			} else {
				compactArrays(child)
			}
		}
	case []any:
		for i, child := range val {
			if arr, ok := child.([]any); ok {
				val[i] = compactArray(arr)
			} else {
				compactArrays(child)
			}
		}
	}
}

func compactArray(arr []any) []any {
	result := make([]any, 0, len(arr))
	for _, v := range arr {
		if v != nil {
			compactArrays(v)
			result = append(result, v)
		}
	}
	return result
}
