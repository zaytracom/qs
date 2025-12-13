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

	// Strip query prefix if requested (match parseValues behavior).
	cleanStr := str
	if opts.IgnoreQueryPrefix && len(cleanStr) > 0 && cleanStr[0] == '?' {
		cleanStr = cleanStr[1:]
	}
	// Decode URL-encoded brackets for easier parsing (match parseValues behavior).
	cleanStr = decodeBrackets(cleanStr)

	// Calculate limit for splitting (match parseValues behavior).
	limit := opts.ParameterLimit
	if opts.ThrowOnLimitExceeded {
		limit = opts.ParameterLimit + 1
	}

	// Split by delimiter (string or regexp), matching parseValues behavior.
	delimiter := opts.Delimiter
	if delimiter == "" {
		delimiter = DefaultDelimiter
	}
	parts := splitByDelimiter(cleanStr, delimiter, opts.DelimiterRegexp, limit)

	// Check parameter limit.
	if opts.ThrowOnLimitExceeded && len(parts) > opts.ParameterLimit {
		return orderedResult{}, ErrParameterLimitExceeded
	}

	// Detect charset from sentinel (same logic as parseValues).
	charset := opts.Charset
	skipIndex := -1
	if opts.CharsetSentinel {
		for i, part := range parts {
			if strings.HasPrefix(part, "utf8=") {
				if part == charsetSentinel {
					charset = CharsetUTF8
					skipIndex = i
				} else if part == isoSentinel {
					charset = CharsetISO88591
					skipIndex = i
				}
				// Unknown utf8=... => treat as regular param (no skip).
				break
			}
		}
	}

	defaultDecoder := func(s string, cs Charset, kind string) (string, error) {
		return Decode(s, cs), nil
	}
	decoder := opts.Decoder
	if decoder == nil {
		decoder = defaultDecoder
	}

	// Parse each part with the AST parser to get bracket-aware key/value split.
	// This works regardless of delimiter type because each part is a single param.
	arena := lang.NewArena(1)
	cfg := lang.DefaultConfig()
	cfg.Delimiter = '&'
	if opts.Comma {
		cfg.Flags |= lang.FlagComma
	}

	for i, part := range parts {
		if i == skipIndex {
			continue
		}

		if part == "" {
			continue
		}

		qsNode, _, err := lang.Parse(arena, part, cfg)
		if err != nil {
			if err == lang.ErrParameterLimitExceeded {
				return orderedResult{}, ErrParameterLimitExceeded
			}
			return orderedResult{}, err
		}

		if qsNode.ParamLen == 0 {
			continue
		}

		p := arena.Params[0]
		keyPart := arena.GetString(p.Key.Raw)

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

			// Handle []= (empty bracket) notation - wrap in array (matches parseValues behavior).
			if strings.Contains(part, "[]=") {
				if arr, ok := val.([]any); ok {
					val = []any{arr}
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
