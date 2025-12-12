# @zaytra/qs-wasm

WebAssembly build of the [QS](https://github.com/zaytracom/qs) query string library. Works in Node.js and browsers.

## Installation

```bash
npm install @zaytra/qs-wasm
```

## Usage

### Node.js

```javascript
import { parse, stringify, createQS } from '@zaytra/qs-wasm';

// Async API
const parsed = await parse('a[b]=c&d=e');
// { a: { b: 'c' }, d: 'e' }

const str = await stringify({ a: { b: 'c' } });
// 'a%5Bb%5D=c'

// Sync API (after initialization)
const qs = await createQS();
const obj = qs.parse('foo=bar');
const query = qs.stringify({ foo: 'bar' });
```

### Browser

```html
<script type="module">
import { parse, stringify, setWasmUrl } from '@zaytra/qs-wasm';

// Optional: set custom path to .wasm file
setWasmUrl('/assets/qs.wasm');

const obj = await parse('a=1&b=2');
const str = await stringify({ foo: 'bar' });
</script>
```

### With bundlers (Vite, Webpack, etc.)

```javascript
import { parse, stringify } from '@zaytra/qs-wasm';

const result = await parse('name=John&age=30');
```

## API

### `parse(queryString, options?): Promise<object>`

Parse a query string into an object.

### `stringify(obj, options?): Promise<string>`

Convert an object to a query string.

### `createQS(): Promise<{ parse, stringify }>`

Create a sync instance (useful for multiple operations).

### `setWasmUrl(url: string): void`

Set custom URL for the .wasm file (browser only, call before using parse/stringify).

## Parse Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `allowDots` | boolean | false | Enable dot notation (`a.b` → `{a:{b:...}}`) |
| `allowEmptyArrays` | boolean | false | Allow `key[]` to be empty array |
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

## Compatibility

This package is compatible with the [qs](https://www.npmjs.com/package/qs) npm package API.

## License

Apache-2.0
