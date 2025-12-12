//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"

	qs "github.com/zaytracom/qs/v2"
)

// parse parses a query string and returns JSON result
func parse(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return jsError("query string required")
	}

	queryString := args[0].String()

	// Parse options from second argument if provided
	var opts []qs.ParseOption
	if len(args) > 1 && !args[1].IsUndefined() && !args[1].IsNull() {
		opts = parseOptionsFromJS(args[1])
	}

	result, err := qs.Parse(queryString, opts...)
	if err != nil {
		return jsError(err.Error())
	}

	return jsResult(result)
}

// stringify converts an object to query string
func stringify(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return js.ValueOf("")
	}

	// Parse JSON input
	var obj map[string]any
	jsonStr := args[0].String()
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return jsError(err.Error())
	}

	// Parse options from second argument if provided
	var opts []qs.StringifyOption
	if len(args) > 1 && !args[1].IsUndefined() && !args[1].IsNull() {
		opts = stringifyOptionsFromJS(args[1])
	}

	result, err := qs.Stringify(obj, opts...)
	if err != nil {
		return jsError(err.Error())
	}

	return js.ValueOf(result)
}

// parseOptionsFromJS converts JS options object to ParseOption slice
func parseOptionsFromJS(jsOpts js.Value) []qs.ParseOption {
	var opts []qs.ParseOption

	if v := jsOpts.Get("allowDots"); !v.IsUndefined() {
		opts = append(opts, qs.WithAllowDots(v.Bool()))
	}
	if v := jsOpts.Get("allowEmptyArrays"); !v.IsUndefined() {
		opts = append(opts, qs.WithAllowEmptyArrays(v.Bool()))
	}
	if v := jsOpts.Get("allowSparse"); !v.IsUndefined() {
		opts = append(opts, qs.WithAllowSparse(v.Bool()))
	}
	if v := jsOpts.Get("arrayLimit"); !v.IsUndefined() {
		opts = append(opts, qs.WithArrayLimit(v.Int()))
	}
	if v := jsOpts.Get("comma"); !v.IsUndefined() {
		opts = append(opts, qs.WithComma(v.Bool()))
	}
	if v := jsOpts.Get("depth"); !v.IsUndefined() {
		opts = append(opts, qs.WithDepth(v.Int()))
	}
	if v := jsOpts.Get("ignoreQueryPrefix"); !v.IsUndefined() {
		opts = append(opts, qs.WithIgnoreQueryPrefix(v.Bool()))
	}
	if v := jsOpts.Get("parameterLimit"); !v.IsUndefined() {
		opts = append(opts, qs.WithParameterLimit(v.Int()))
	}
	if v := jsOpts.Get("parseArrays"); !v.IsUndefined() {
		opts = append(opts, qs.WithParseArrays(v.Bool()))
	}
	if v := jsOpts.Get("strictNullHandling"); !v.IsUndefined() {
		opts = append(opts, qs.WithStrictNullHandling(v.Bool()))
	}
	if v := jsOpts.Get("delimiter"); !v.IsUndefined() {
		opts = append(opts, qs.WithDelimiter(v.String()))
	}

	return opts
}

// stringifyOptionsFromJS converts JS options object to StringifyOption slice
func stringifyOptionsFromJS(jsOpts js.Value) []qs.StringifyOption {
	var opts []qs.StringifyOption

	if v := jsOpts.Get("addQueryPrefix"); !v.IsUndefined() {
		opts = append(opts, qs.WithStringifyAddQueryPrefix(v.Bool()))
	}
	if v := jsOpts.Get("allowDots"); !v.IsUndefined() {
		opts = append(opts, qs.WithStringifyAllowDots(v.Bool()))
	}
	if v := jsOpts.Get("allowEmptyArrays"); !v.IsUndefined() {
		opts = append(opts, qs.WithStringifyAllowEmptyArrays(v.Bool()))
	}
	if v := jsOpts.Get("arrayFormat"); !v.IsUndefined() {
		switch v.String() {
		case "indices":
			opts = append(opts, qs.WithArrayFormat(qs.ArrayFormatIndices))
		case "brackets":
			opts = append(opts, qs.WithArrayFormat(qs.ArrayFormatBrackets))
		case "repeat":
			opts = append(opts, qs.WithArrayFormat(qs.ArrayFormatRepeat))
		case "comma":
			opts = append(opts, qs.WithArrayFormat(qs.ArrayFormatComma))
		}
	}
	if v := jsOpts.Get("encode"); !v.IsUndefined() {
		opts = append(opts, qs.WithEncode(v.Bool()))
	}
	if v := jsOpts.Get("encodeValuesOnly"); !v.IsUndefined() {
		opts = append(opts, qs.WithEncodeValuesOnly(v.Bool()))
	}
	if v := jsOpts.Get("skipNulls"); !v.IsUndefined() {
		opts = append(opts, qs.WithSkipNulls(v.Bool()))
	}
	if v := jsOpts.Get("strictNullHandling"); !v.IsUndefined() {
		opts = append(opts, qs.WithStringifyStrictNullHandling(v.Bool()))
	}
	if v := jsOpts.Get("delimiter"); !v.IsUndefined() {
		opts = append(opts, qs.WithStringifyDelimiter(v.String()))
	}

	return opts
}

// jsError returns a JS object with error field
func jsError(msg string) js.Value {
	return js.ValueOf(map[string]any{
		"error": msg,
	})
}

// jsResult returns a JS object with result field (JSON string)
func jsResult(data any) js.Value {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return jsError(err.Error())
	}
	return js.ValueOf(map[string]any{
		"result": string(jsonBytes),
	})
}

func main() {
	// Export functions to global scope
	js.Global().Set("qsParse", js.FuncOf(parse))
	js.Global().Set("qsStringify", js.FuncOf(stringify))

	// Keep the program running
	<-make(chan struct{})
}
