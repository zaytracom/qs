# QS WASM Module

WebAssembly build of the QS library for use in Node.js and browsers.

## Quick Start

```bash
# Build WASM module
make build

# Run tests
make test
```

## Usage in Node.js

```javascript
import { createQS } from './index.mjs';

const qs = await createQS();

// Parse query string
const parsed = qs.parse('a[b]=c&d=e');
// { a: { b: 'c' }, d: 'e' }

// With options
const dotParsed = qs.parse('a.b.c=d', { allowDots: true });
// { a: { b: { c: 'd' } } }

// Stringify object
const str = qs.stringify({ a: { b: 'c' } });
// 'a%5Bb%5D=c'

// With options
const dotStr = qs.stringify({ a: { b: 'c' } }, { allowDots: true, encode: false });
// 'a.b=c'
```

## Parse Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `allowDots` | boolean | false | Enable dot notation (`a.b` → `{a:{b:...}}`) |
| `allowEmptyArrays` | boolean | false | Allow `key[]` with empty value to be `[]` |
| `allowSparse` | boolean | false | Preserve sparse arrays |
| `arrayLimit` | number | 20 | Max index for array parsing |
| `comma` | boolean | false | Parse comma-separated values |
| `depth` | number | 5 | Max nesting depth |
| `ignoreQueryPrefix` | boolean | false | Strip leading `?` |
| `parameterLimit` | number | 1000 | Max parameters to parse |
| `parseArrays` | boolean | true | Enable array parsing |
| `strictNullHandling` | boolean | false | Treat `key` (no value) as `null` |
| `delimiter` | string | `&` | Parameter delimiter |

## Stringify Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `addQueryPrefix` | boolean | false | Add leading `?` |
| `allowDots` | boolean | false | Use dot notation |
| `allowEmptyArrays` | boolean | false | Output `key[]` for empty arrays |
| `arrayFormat` | string | `indices` | `indices`, `brackets`, `repeat`, or `comma` |
| `encode` | boolean | true | URL encode output |
| `encodeValuesOnly` | boolean | false | Only encode values, not keys |
| `skipNulls` | boolean | false | Skip null/undefined values |
| `strictNullHandling` | boolean | false | Omit `=` for null values |
| `delimiter` | string | `&` | Parameter delimiter |

## Build Commands

```bash
make build           # Build qs.wasm
make build-optimized # Build smaller qs.wasm (stripped)
make test            # Build and run tests
make clean           # Remove qs.wasm
make update-wasm-exec # Update wasm_exec.js from Go
```
