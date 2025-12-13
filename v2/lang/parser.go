package lang

const (
	maxUint8  = ^uint8(0)
	maxUint16 = ^uint16(0)
	maxUint32 = ^uint32(0)
)

const (
	noValue = uint16(0xFFFF)
)

var (
	utf8SentinelValue = "%E2%9C%93"
	isoSentinelValue  = "%26%2310003%3B"
)

// Parser parses a query string into an Arena-backed AST.
// Reuse both Parser and Arena to keep parsing allocation-free.
type Parser struct {
	cfg Config

	arena *Arena
	src   string

	// internal
	sentinelChecked bool
	detectedCharset Charset
}

func (p *Parser) Reset(arena *Arena, cfg Config) {
	p.cfg = cfg
	if p.cfg.Delimiter == 0 {
		p.cfg.Delimiter = DefaultDelimiter
	}
	p.arena = arena
	p.src = ""
	p.sentinelChecked = false
	p.detectedCharset = p.cfg.Charset
}

// ParseInto parses input into the provided arena and returns the root AST node.
func (p *Parser) ParseInto(input string) (QueryString, Charset, error) {
	if p.arena == nil {
		return QueryString{}, p.cfg.Charset, ErrNilArena
	}

	p.arena.Reset(input)
	p.src = p.arena.Source
	p.sentinelChecked = false
	p.detectedCharset = p.cfg.Charset

	hasPrefix := len(p.src) > 0 && p.src[0] == '?'
	start := 0
	if hasPrefix && p.cfg.Flags.Has(FlagIgnoreQueryPrefix) {
		start = 1
	}

	qs := QueryString{
		HasPrefix:  hasPrefix,
		ParamStart: 0,
		ParamLen:   0,
	}

	if start >= len(p.src) {
		return qs, p.detectedCharset, nil
	}

	var (
		stateKey   = uint8(0)
		stateValue = uint8(1)
	)

	state := stateKey
	paramStart := uint32(start)
	keyEnd := uint32(start)
	hasEquals := false
	bracketDepth := uint16(0)
	firstIgnoredEquals := uint32(0)

	var emitted uint32
	limit := uint32(p.cfg.ParameterLimit)
	if p.cfg.ParameterLimit == 0 {
		limit = maxUint32
	}

	srcLen := uint32(len(p.src))
	for i := uint32(start); i <= srcLen; i++ {
		var c byte
		if i < srcLen {
			c = p.src[i]
		}

		if i == srcLen || c == p.cfg.Delimiter {
			end := i
			if state == stateKey {
				keyEnd = end
				hasEquals = false
				if firstIgnoredEquals != 0 {
					keyEnd = firstIgnoredEquals
					hasEquals = true
				}
			} else {
				hasEquals = true
			}

			if end > paramStart || hasEquals {
				if emitted >= limit {
					if p.cfg.Flags.Has(FlagThrowOnLimitExceeded) {
						return QueryString{}, p.detectedCharset, ErrParameterLimitExceeded
					}
					break
				}
				before := len(p.arena.Params)
				if err := p.emitParam(paramStart, keyEnd, end, hasEquals); err != nil {
					return QueryString{}, p.detectedCharset, err
				}
				if len(p.arena.Params) != before {
					emitted++
				}
			}

			if i == srcLen {
				break
			}
			paramStart = i + 1
			state = stateKey
			keyEnd = paramStart
			hasEquals = false
			bracketDepth = 0
			firstIgnoredEquals = 0
			continue
		}

		if state == stateKey {
			ii := int(i)
			end := int(srcLen)

			if openLen := lbracketTokenLen(p.src, ii, end); openLen != 0 {
				bracketDepth++
				i += uint32(openLen - 1)
				continue
			}
			if closeLen := rbracketTokenLen(p.src, ii, end); closeLen != 0 {
				if bracketDepth > 0 {
					bracketDepth--
					if bracketDepth == 0 {
						next := i + uint32(closeLen)
						if next < srcLen && p.src[next] == '=' {
							keyEnd = next
							state = stateValue
							i = next
							continue
						}
					}
				}
				i += uint32(closeLen - 1)
				continue
			}

			if c == '=' {
				if bracketDepth == 0 {
					keyEnd = i
					state = stateValue
				} else if firstIgnoredEquals == 0 {
					firstIgnoredEquals = i
				}
			}
		}
	}

	if len(p.arena.Params) > int(maxUint16) {
		return QueryString{}, p.detectedCharset, ErrTooManyParams
	}
	qs.ParamLen = uint16(len(p.arena.Params))
	return qs, p.detectedCharset, nil
}

// Parse parses input with cfg into arena and returns the AST.
func Parse(arena *Arena, input string, cfg Config) (QueryString, Charset, error) {
	var p Parser
	p.Reset(arena, cfg)
	return p.ParseInto(input)
}

func (p *Parser) emitParam(paramStart, keyEnd, paramEnd uint32, hasEquals bool) error {
	// snapshot for rollback on skip/error
	paramsN := len(p.arena.Params)
	segsN := len(p.arena.Segments)
	valsN := len(p.arena.Values)
	partsN := len(p.arena.ValueParts)

	// empty key => ignore (matches qs behavior)
	if keyEnd <= paramStart {
		return nil
	}

	keySpan, err := p.makeSpan(paramStart, keyEnd)
	if err != nil {
		return err
	}

	// Charset sentinel (only check first "utf8=" occurrence)
	if hasEquals && p.cfg.Flags.Has(FlagCharsetSentinel) && !p.sentinelChecked &&
		spanEqualsASCII(p.src, keySpan, "utf8") {
		p.sentinelChecked = true

		valSpan, err := p.makeSpan(keyEnd+1, paramEnd)
		if err != nil {
			return err
		}

		if spanEqualsFoldASCII(p.src, valSpan, utf8SentinelValue) {
			p.detectedCharset = CharsetUTF8
			return nil
		}
		if spanEqualsFoldASCII(p.src, valSpan, isoSentinelValue) {
			p.detectedCharset = CharsetISO88591
			return nil
		}
		// unknown utf8=... => treat as regular parameter
	}

	if p.cfg.Profile == ProfileStrict {
		if err := validatePercentEncoding(p.src, keySpan); err != nil {
			return err
		}
	}

	key, skip, err := p.parseKey(paramStart, keyEnd)
	if err != nil {
		p.rollback(paramsN, segsN, valsN, partsN)
		return err
	}
	if skip {
		p.rollback(paramsN, segsN, valsN, partsN)
		return nil
	}

	param := Param{
		Key:       key,
		ValueIdx:  noValue,
		HasEquals: hasEquals,
	}

	if hasEquals {
		valIdx, err := p.emitValue(keyEnd+1, paramEnd)
		if err != nil {
			p.rollback(paramsN, segsN, valsN, partsN)
			return err
		}
		param.ValueIdx = valIdx
	} else if p.cfg.Flags.Has(FlagStrictNullHandling) {
		valIdx, err := p.emitNullValue(paramEnd)
		if err != nil {
			p.rollback(paramsN, segsN, valsN, partsN)
			return err
		}
		param.ValueIdx = valIdx
	}

	if len(p.arena.Params) >= int(maxUint16) {
		p.rollback(paramsN, segsN, valsN, partsN)
		return ErrTooManyParams
	}
	p.arena.Params = append(p.arena.Params, param)
	return nil
}

func (p *Parser) rollback(paramsN, segsN, valsN, partsN int) {
	p.arena.Params = p.arena.Params[:paramsN]
	p.arena.Segments = p.arena.Segments[:segsN]
	p.arena.Values = p.arena.Values[:valsN]
	p.arena.ValueParts = p.arena.ValueParts[:partsN]
}

func (p *Parser) emitNullValue(at uint32) (uint16, error) {
	idx := len(p.arena.Values)
	if idx > int(maxUint16) {
		return 0, ErrTooManyValues
	}
	span, err := p.makeSpan(at, at)
	if err != nil {
		return 0, err
	}
	p.arena.Values = append(p.arena.Values, Value{
		Kind: ValNull,
		Raw:  span,
	})
	return uint16(idx), nil
}

func (p *Parser) emitValue(valStart, valEnd uint32) (uint16, error) {
	idx := len(p.arena.Values)
	if idx > int(maxUint16) {
		return 0, ErrTooManyValues
	}

	raw, err := p.makeSpan(valStart, valEnd)
	if err != nil {
		return 0, err
	}

	if p.cfg.Profile == ProfileStrict {
		if err := validatePercentEncoding(p.src, raw); err != nil {
			return 0, err
		}
	}

	if !p.cfg.Flags.Has(FlagComma) {
		p.arena.Values = append(p.arena.Values, Value{Kind: ValSimple, Raw: raw})
		return uint16(idx), nil
	}

	// fast scan: split only if we find an actual comma
	if !spanContainsByte(p.src, raw, ',') {
		p.arena.Values = append(p.arena.Values, Value{Kind: ValSimple, Raw: raw})
		return uint16(idx), nil
	}

	partsOff := len(p.arena.ValueParts)
	if partsOff > int(maxUint16) {
		return 0, ErrTooManyValueParts
	}

	partStart := valStart
	var partsLen int
	for i := valStart; i <= valEnd; i++ {
		if i == valEnd || p.src[i] == ',' {
			part, err := p.makeSpan(partStart, i)
			if err != nil {
				return 0, err
			}
			if len(p.arena.ValueParts) > int(maxUint16) {
				return 0, ErrTooManyValueParts
			}
			p.arena.ValueParts = append(p.arena.ValueParts, part)
			partsLen++
			partStart = i + 1
		}
	}
	if partsLen > int(maxUint8) {
		return 0, ErrTooManyValueParts
	}

	p.arena.Values = append(p.arena.Values, Value{
		Kind:     ValComma,
		Raw:      raw,
		PartsOff: uint16(partsOff),
		PartsLen: uint8(partsLen),
	})
	return uint16(idx), nil
}

func (p *Parser) parseKey(keyStart, keyEnd uint32) (Key, bool, error) {
	segBase := len(p.arena.Segments)

	raw, err := p.makeSpan(keyStart, keyEnd)
	if err != nil {
		return Key{}, false, err
	}

	start := int(keyStart)
	end := int(keyEnd)

	// Flat mode: treat entire key as a single segment.
	if p.cfg.Depth == 0 {
		seg, skip, err := p.makeIdentSegment(raw, NotationRoot, segBase == len(p.arena.Segments))
		if err != nil {
			return Key{}, false, err
		}
		if skip {
			return Key{}, true, nil
		}
		if err := p.appendSegment(seg); err != nil {
			return Key{}, false, err
		}
		return p.finishKey(segBase, raw)
	}

	segStart := start
	segNotation := NotationRoot

	var depthUsed uint16

	for i := start; i < end; {
		if openLen := lbracketTokenLen(p.src, i, end); openLen != 0 {
			contentStart := i + openLen
			closePos, closeLen, ok := findBracketClose(p.src, contentStart, end)
			if ok {
				// Emit preceding identifier segment (if any) before '['.
				if segStart < i {
					segSpan, err := p.makeSpan(uint32(segStart), uint32(i))
					if err != nil {
						p.arena.Segments = p.arena.Segments[:segBase]
						return Key{}, false, err
					}
					seg, skip, err := p.makeIdentSegment(segSpan, segNotation, segBase == len(p.arena.Segments))
					if err != nil {
						p.arena.Segments = p.arena.Segments[:segBase]
						return Key{}, false, err
					}
					if skip {
						p.arena.Segments = p.arena.Segments[:segBase]
						return Key{}, true, nil
					}
					if err := p.appendSegment(seg); err != nil {
						p.arena.Segments = p.arena.Segments[:segBase]
						return Key{}, false, err
					}
					if !(seg.Notation == NotationRoot && len(p.arena.Segments)-segBase == 1) {
						depthUsed++
					}
				}

				// Depth limit check for the bracket segment itself.
				if p.cfg.Depth > 0 && depthUsed >= p.cfg.Depth {
					if p.cfg.Flags.Has(FlagStrictDepth) {
						p.arena.Segments = p.arena.Segments[:segBase]
						return Key{}, false, ErrDepthLimitExceeded
					}
					litSpan, err := p.makeSpan(uint32(i), keyEnd)
					if err != nil {
						p.arena.Segments = p.arena.Segments[:segBase]
						return Key{}, false, err
					}
					if err := p.appendSegment(Segment{
						Kind:     SegLiteral,
						Notation: NotationRoot, Span: litSpan, Index: -1,
					}); err != nil {
						p.arena.Segments = p.arena.Segments[:segBase]
						return Key{}, false, err
					}
					return p.finishKey(segBase, raw)
				}

				contentSpan, err := p.makeSpan(uint32(contentStart), uint32(closePos))
				if err != nil {
					p.arena.Segments = p.arena.Segments[:segBase]
					return Key{}, false, err
				}
				seg, skip, err := p.makeBracketSegment(contentSpan)
				if err != nil {
					p.arena.Segments = p.arena.Segments[:segBase]
					return Key{}, false, err
				}
				if skip {
					p.arena.Segments = p.arena.Segments[:segBase]
					return Key{}, true, nil
				}
				if err := p.appendSegment(seg); err != nil {
					p.arena.Segments = p.arena.Segments[:segBase]
					return Key{}, false, err
				}
				depthUsed++

				segStart = closePos + closeLen
				segNotation = NotationRoot
				i = segStart
				continue
			}
			i += openLen
			continue
		}

		if p.cfg.Flags.Has(FlagAllowDots) {
			if dotLen := dotTokenLen(p.src, i, end,
				p.cfg.Flags.Has(FlagDecodeDotInKeys)); dotLen != 0 {
				next := i + dotLen
				if next < end && p.src[next] != '.' &&
					lbracketTokenLen(p.src, next, end) == 0 {
					// Emit current identifier segment before dot, if any.
					if segStart < i {
						segSpan, err := p.makeSpan(uint32(segStart), uint32(i))
						if err != nil {
							p.arena.Segments = p.arena.Segments[:segBase]
							return Key{}, false, err
						}
						seg, skip, err := p.makeIdentSegment(segSpan,
							segNotation, segBase == len(p.arena.Segments))
						if err != nil {
							p.arena.Segments = p.arena.Segments[:segBase]
							return Key{}, false, err
						}
						if skip {
							p.arena.Segments = p.arena.Segments[:segBase]
							return Key{}, true, nil
						}
						if err := p.appendSegment(seg); err != nil {
							p.arena.Segments = p.arena.Segments[:segBase]
							return Key{}, false, err
						}
						if !(seg.Notation == NotationRoot && len(p.arena.Segments)-segBase == 1) {
							depthUsed++
						}
					} else if len(p.arena.Segments) == segBase && i != start {
						// No preceding segment to attach the dot to.
						i++
						continue
					}

					// Depth limit check for the segment after dot.
					if p.cfg.Depth > 0 && depthUsed >= p.cfg.Depth {
						if p.cfg.Flags.Has(FlagStrictDepth) {
							p.arena.Segments = p.arena.Segments[:segBase]
							return Key{}, false, ErrDepthLimitExceeded
						}
						litSpan, err := p.makeSpan(uint32(i), keyEnd)
						if err != nil {
							p.arena.Segments = p.arena.Segments[:segBase]
							return Key{}, false, err
						}
						if err := p.appendSegment(Segment{
							Kind:     SegLiteral,
							Notation: NotationRoot, Span: litSpan, Index: -1,
						}); err != nil {
							p.arena.Segments = p.arena.Segments[:segBase]
							return Key{}, false, err
						}
						return p.finishKey(segBase, raw)
					}

					segStart = next
					segNotation = NotationDot
					i = next
					continue
				}
			}
		}

		i++
	}

	if segStart < end {
		segSpan, err := p.makeSpan(uint32(segStart), keyEnd)
		if err != nil {
			p.arena.Segments = p.arena.Segments[:segBase]
			return Key{}, false, err
		}
		seg, skip, err := p.makeIdentSegment(segSpan, segNotation, segBase == len(p.arena.Segments))
		if err != nil {
			p.arena.Segments = p.arena.Segments[:segBase]
			return Key{}, false, err
		}
		if skip {
			p.arena.Segments = p.arena.Segments[:segBase]
			return Key{}, true, nil
		}
		if err := p.appendSegment(seg); err != nil {
			p.arena.Segments = p.arena.Segments[:segBase]
			return Key{}, false, err
		}
	}

	return p.finishKey(segBase, raw)
}

func (p *Parser) finishKey(segBase int, raw Span) (Key, bool, error) {
	segStart := segBase
	segLen := len(p.arena.Segments) - segBase

	if segLen == 0 {
		return Key{}, true, nil
	}
	if segStart > int(maxUint16) {
		return Key{}, false, ErrTooManySegments
	}
	if segLen > int(maxUint8) {
		return Key{}, false, ErrTooManySegments
	}
	return Key{
		SegStart: uint16(segStart),
		SegLen:   uint8(segLen),
		Raw:      raw,
	}, false, nil
}

func (p *Parser) appendSegment(seg Segment) error {
	if len(p.arena.Segments) > int(maxUint16) {
		return ErrTooManySegments
	}
	p.arena.Segments = append(p.arena.Segments, seg)
	return nil
}

func (p *Parser) makeBracketSegment(content Span) (Segment, bool, error) {
	// [] -> SegEmpty
	if content.Len == 0 {
		seg := Segment{
			Kind:     SegEmpty,
			Notation: NotationBracket,
			Span:     content,
			Index:    -1,
		}
		if p.isBlockedSegment(seg) {
			return Segment{}, true, nil
		}
		return seg, false, nil
	}

	kind := SegIdent
	index := int32(-1)
	if p.cfg.ParseArrays {
		if v, ok := parseCanonicalUint(content, p.src); ok {
			if v <= uint32(p.cfg.ArrayLimit) {
				kind = SegIndex
				index = int32(v)
			}
			// v > ArrayLimit: keep as SegIdent to be treated as an object key ("{ '21': ... }").
		}
	}
	seg := Segment{
		Kind:     kind,
		Notation: NotationBracket,
		Span:     content,
		Index:    index,
	}
	if p.isBlockedSegment(seg) {
		return Segment{}, true, nil
	}
	return seg, false, nil
}

func (p *Parser) makeIdentSegment(span Span, notation Notation, isFirst bool) (Segment, bool, error) {
	kind := SegIdent
	index := int32(-1)

	// Only non-root segments can be indexes (qs compatibility).
	if notation != NotationRoot && p.cfg.ParseArrays {
		if v, ok := parseCanonicalUint(span, p.src); ok {
			if v <= uint32(p.cfg.ArrayLimit) {
				kind = SegIndex
				index = int32(v)
			}
		}
	}

	seg := Segment{
		Kind:     kind,
		Notation: notation,
		Span:     span,
		Index:    index,
	}
	_ = isFirst // reserved for future semantics; currently unused

	if p.isBlockedSegment(seg) {
		return Segment{}, true, nil
	}
	return seg, false, nil
}

func (p *Parser) isBlockedSegment(seg Segment) bool {
	// In Go there's no prototype pollution, so nothing is blocked
	return false
}

func (p *Parser) makeSpan(start, end uint32) (Span, error) {
	if end < start {
		return Span{}, ErrSpanTooLarge
	}
	if end > uint32(len(p.src)) {
		return Span{}, ErrSpanTooLarge
	}
	n := end - start
	if n > uint32(maxUint16) {
		return Span{}, ErrSpanTooLarge
	}
	return Span{Off: start, Len: uint16(n)}, nil
}

func spanContainsByte(src string, sp Span, b byte) bool {
	start := int(sp.Off)
	end := start + int(sp.Len)
	for i := start; i < end; i++ {
		if src[i] == b {
			return true
		}
	}
	return false
}

func lbracketTokenLen(src string, i, end int) int {
	if i >= end {
		return 0
	}
	if src[i] == '[' {
		return 1
	}
	if src[i] == '%' && i+2 < end && src[i+1] == '5' {
		if b := src[i+2]; b == 'B' || b == 'b' {
			return 3
		}
	}
	return 0
}

func rbracketTokenLen(src string, i, end int) int {
	if i >= end {
		return 0
	}
	if src[i] == ']' {
		return 1
	}
	if src[i] == '%' && i+2 < end && src[i+1] == '5' {
		if b := src[i+2]; b == 'D' || b == 'd' {
			return 3
		}
	}
	return 0
}

func dotTokenLen(src string, i, end int, allowEncoded bool) int {
	if i >= end {
		return 0
	}
	if src[i] == '.' {
		return 1
	}
	if allowEncoded && src[i] == '%' && i+2 < end && src[i+1] == '2' {
		if b := src[i+2]; b == 'E' || b == 'e' {
			return 3
		}
	}
	return 0
}

func findBracketClose(src string, contentStart, end int) (pos, closeLen int, ok bool) {
	for i := contentStart; i < end; i++ {
		if lbracketTokenLen(src, i, end) != 0 {
			return 0, 0, false
		}
		if closeLen = rbracketTokenLen(src, i, end); closeLen != 0 {
			return i, closeLen, true
		}
	}
	return 0, 0, false
}

func parseCanonicalUint(sp Span, src string) (uint32, bool) {
	if sp.Len == 0 {
		return 0, false
	}
	start := int(sp.Off)
	end := start + int(sp.Len)
	if start < 0 || end > len(src) {
		return 0, false
	}

	if src[start] == '0' && sp.Len > 1 {
		return 0, false
	}

	var n uint32
	for i := start; i < end; i++ {
		c := src[i]
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + uint32(c-'0')
	}
	return n, true
}

func spanEqualsASCII(src string, sp Span, s string) bool {
	if int(sp.Len) != len(s) {
		return false
	}
	start := int(sp.Off)
	end := start + int(sp.Len)
	if start < 0 || end > len(src) {
		return false
	}
	for i := 0; i < len(s); i++ {
		if src[start+i] != s[i] {
			return false
		}
	}
	return true
}

func spanEqualsFoldASCII(src string, sp Span, s string) bool {
	if int(sp.Len) != len(s) {
		return false
	}
	start := int(sp.Off)
	end := start + int(sp.Len)
	if start < 0 || end > len(src) {
		return false
	}
	for i := 0; i < len(s); i++ {
		a := src[start+i]
		b := s[i]
		if a == b {
			continue
		}
		if toLowerASCII(a) != toLowerASCII(b) {
			return false
		}
	}
	return true
}

func spanEqualsDecodedASCII(src string, sp Span, s string) bool {
	start := int(sp.Off)
	end := start + int(sp.Len)
	if start < 0 || end > len(src) {
		return false
	}

	j := 0
	for i := start; i < end; i++ {
		if j >= len(s) {
			return false
		}
		c := src[i]
		switch c {
		case '+':
			if s[j] != ' ' {
				return false
			}
			j++
		case '%':
			if i+2 >= end {
				return false
			}
			hi := fromHex(src[i+1])
			lo := fromHex(src[i+2])
			if hi < 0 || lo < 0 {
				return false
			}
			if byte(hi<<4|lo) != s[j] {
				return false
			}
			i += 2
			j++
		default:
			if c != s[j] {
				return false
			}
			j++
		}
	}
	return j == len(s)
}

func validatePercentEncoding(src string, sp Span) error {
	start := int(sp.Off)
	end := start + int(sp.Len)
	if start < 0 || end > len(src) {
		return ErrSpanTooLarge
	}
	for i := start; i < end; i++ {
		if src[i] != '%' {
			continue
		}
		if i+2 >= end {
			return ErrInvalidPercentCode
		}
		if fromHex(src[i+1]) < 0 || fromHex(src[i+2]) < 0 {
			return ErrInvalidPercentCode
		}
		i += 2
	}
	return nil
}

func fromHex(b byte) int {
	switch {
	case b >= '0' && b <= '9':
		return int(b - '0')
	case b >= 'a' && b <= 'f':
		return int(b-'a') + 10
	case b >= 'A' && b <= 'F':
		return int(b-'A') + 10
	default:
		return -1
	}
}

func toLowerASCII(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}
