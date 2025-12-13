package qs

import (
	"strings"

	"github.com/zaytracom/qs/v2/lang"
)

func parseValuesAST(str string, opts *ParseOptions) (orderedResult, error) {
	result := orderedResult{
		keys:   make([]string, 0),
		values: make(map[string]any),
	}

	// We only support single-byte delimiters here; callers must fallback otherwise.
	delimiter := opts.Delimiter
	if delimiter == "" {
		delimiter = DefaultDelimiter
	}

	cfg := lang.DefaultConfig()
	cfg.Delimiter = delimiter[0]
	cfg.PlainObjects = true // don't block prototype keys at AST layer
	if opts.IgnoreQueryPrefix {
		cfg.Flags |= lang.FlagIgnoreQueryPrefix
	}
	if opts.Comma {
		cfg.Flags |= lang.FlagComma
	}
	if opts.ThrowOnLimitExceeded {
		cfg.Flags |= lang.FlagThrowOnLimitExceeded
	}

	if opts.ParameterLimit <= 0 {
		cfg.ParameterLimit = 0 // unlimited
	} else if opts.ParameterLimit > int(^uint16(0)) {
		cfg.ParameterLimit = ^uint16(0)
	} else {
		cfg.ParameterLimit = uint16(opts.ParameterLimit)
	}

	arena := lang.NewArena(estimateParams(str))
	qsNode, _, err := lang.Parse(arena, str, cfg)
	if err != nil {
		if err == lang.ErrParameterLimitExceeded {
			return orderedResult{}, ErrParameterLimitExceeded
		}
		return orderedResult{}, err
	}

	// Detect charset from sentinel, matching parseValues behavior.
	charset := opts.Charset
	skipIndex := -1
	if opts.CharsetSentinel {
		for i := uint16(0); i < qsNode.ParamLen; i++ {
			p := arena.Params[qsNode.ParamStart+i]
			if !p.HasEquals {
				continue
			}
			keyRaw := arena.GetString(p.Key.Raw)
			if keyRaw != "utf8" {
				continue
			}
			v := arena.Values[p.ValueIdx]
			valRaw := arena.GetString(v.Raw)
			if valRaw == strings.TrimPrefix(charsetSentinel, "utf8=") {
				charset = CharsetUTF8
				skipIndex = int(i)
			} else if valRaw == strings.TrimPrefix(isoSentinel, "utf8=") {
				charset = CharsetISO88591
				skipIndex = int(i)
			}
			// Unknown utf8=... => keep as regular param (no skip).
			break
		}
	}

	defaultDecoder := func(s string, cs Charset, kind string) (string, error) {
		return Decode(s, cs), nil
	}
	decoder := opts.Decoder
	if decoder == nil {
		decoder = defaultDecoder
	}

	for i := uint16(0); i < qsNode.ParamLen; i++ {
		if int(i) == skipIndex {
			continue
		}

		p := arena.Params[qsNode.ParamStart+i]
		keyPart := arena.GetString(p.Key.Raw)
		if keyPart == "" {
			continue
		}

		decodedKey, err := decoder(keyPart, charset, "key")
		if err != nil {
			return orderedResult{}, err
		}
		if decodedKey == "" {
			continue
		}

		var val any
		if !p.HasEquals {
			if opts.StrictNullHandling {
				val = ExplicitNullValue
			} else {
				val = ""
			}
		} else {
			// Get current array length for limit checking (matches parseValues).
			currentLen := 0
			if existing, ok := result.values[decodedKey]; ok {
				if arr, isArr := existing.([]any); isArr {
					currentLen = len(arr)
				}
			}

			rawVal := ""
			if p.ValueIdx != uint16(0xFFFF) {
				rawVal = arena.GetString(arena.Values[p.ValueIdx].Raw)
			}

			parsedVal, err := parseArrayValue(rawVal, opts, currentLen)
			if err != nil {
				return orderedResult{}, err
			}

			// If comma flag is enabled and the AST tokenizer already split, use parts.
			if opts.Comma && p.ValueIdx != uint16(0xFFFF) {
				v := arena.Values[p.ValueIdx]
				if v.Kind == lang.ValComma {
					parts := make([]any, 0, int(v.PartsLen))
					for j := uint16(0); j < uint16(v.PartsLen); j++ {
						partRaw := arena.GetString(arena.ValueParts[v.PartsOff+j])
						decoded, err := decoder(partRaw, charset, "value")
						if err != nil {
							return orderedResult{}, err
						}
						parts = append(parts, decoded)
					}
					val = parts
				}
			}

			if val == nil {
				// Decode parsedVal (string or []any of strings) like parseValues.
				if slice, ok := parsedVal.([]any); ok {
					decodedSlice := make([]any, len(slice))
					for i, v := range slice {
						if s, ok := v.(string); ok {
							decoded, err := decoder(s, charset, "value")
							if err != nil {
								return orderedResult{}, err
							}
							decodedSlice[i] = decoded
						} else {
							decodedSlice[i] = v
						}
					}
					val = decodedSlice
				} else if s, ok := parsedVal.(string); ok {
					decoded, err := decoder(s, charset, "value")
					if err != nil {
						return orderedResult{}, err
					}
					val = decoded
				} else {
					val = parsedVal
				}
			}

			// Interpret numeric entities if enabled
			if val != nil && opts.InterpretNumericEntities && charset == CharsetISO88591 {
				if s, ok := val.(string); ok {
					val = interpretNumericEntitiesFunc(s)
				} else if arr, ok := val.([]any); ok {
					for i, v := range arr {
						if s, ok := v.(string); ok {
							arr[i] = interpretNumericEntitiesFunc(s)
						}
					}
				}
			}

			// Handle []= (empty bracket) notation - wrap in array to preserve nested arrays.
			if p.Key.SegLen > 0 {
				last := arena.Segments[p.Key.SegStart+uint16(p.Key.SegLen)-1]
				if last.Kind == lang.SegEmpty {
					if arr, ok := val.([]any); ok {
						val = []any{arr}
					}
				}
			}
		}

		if existing, exists := result.values[decodedKey]; exists {
			switch opts.Duplicates {
			case DuplicateCombine:
				result.values[decodedKey] = Combine(existing, val)
			case DuplicateFirst:
				// keep existing
			case DuplicateLast:
				result.values[decodedKey] = val
			default:
				result.values[decodedKey] = Combine(existing, val)
			}
		} else {
			result.keys = append(result.keys, decodedKey)
			result.values[decodedKey] = val
		}
	}

	return result, nil
}

func estimateParams(s string) int {
	// Cheap estimate: 1 + number of delimiters.
	if s == "" {
		return 0
	}
	n := 1
	for i := 0; i < len(s); i++ {
		if s[i] == '&' {
			n++
		}
	}
	return n
}
