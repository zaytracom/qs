/**
 * JS Compatibility Tests for QS WASM Module
 *
 * Compares the WASM implementation against the original Node.js qs library
 * to ensure compatibility.
 */

import { parse, stringify, createQS } from '../dist/index.js';
import qsOriginal from 'qs';
import assert from 'assert';

// Test counter
let passed = 0;
let failed = 0;

async function test(name, fn) {
	try {
		await fn();
		passed++;
		console.log(`✓ ${name}`);
	} catch (err) {
		failed++;
		console.error(`✗ ${name}`);
		console.error(`  ${err.message}`);
	}
}

function deepEqual(a, b) {
	return JSON.stringify(sortObject(a)) === JSON.stringify(sortObject(b));
}

function sortObject(obj) {
	if (obj === null || typeof obj !== 'object') return obj;
	if (Array.isArray(obj)) return obj.map(sortObject);
	const sorted = {};
	for (const key of Object.keys(obj).sort()) {
		sorted[key] = sortObject(obj[key]);
	}
	return sorted;
}

function compareQueryStrings(a, b) {
	const parseQS = (s) => {
		const result = {};
		if (!s) return result;
		s = s.replace(/^\?/, '');
		for (const part of s.split('&')) {
			const [key, val = ''] = part.split('=');
			if (!result[key]) result[key] = [];
			result[key].push(val);
		}
		for (const k of Object.keys(result)) {
			result[k].sort();
		}
		return result;
	};
	return deepEqual(parseQS(a), parseQS(b));
}

console.log('=== QS WASM vs Node.js qs Compatibility Tests ===\n');

// Initialize WASM
const qs = await createQS();

// =============================================================================
// PARSE TESTS
// =============================================================================

console.log('\n--- Parse Tests ---\n');

// Test 1: Basic parse
await test('Parse: basic', async () => {
	const input = 'a=b&c=d';
	const wasmResult = await parse(input);
	const jsResult = qsOriginal.parse(input);
	assert(deepEqual(wasmResult, jsResult), `WASM: ${JSON.stringify(wasmResult)}, JS: ${JSON.stringify(jsResult)}`);
});

// Test 2: Nested objects
await test('Parse: nested objects', async () => {
	const input = 'a[b][c]=d';
	const wasmResult = await parse(input);
	const jsResult = qsOriginal.parse(input);
	assert(deepEqual(wasmResult, jsResult), `WASM: ${JSON.stringify(wasmResult)}, JS: ${JSON.stringify(jsResult)}`);
});

// Test 3: Arrays
await test('Parse: arrays', async () => {
	const input = 'a[0]=b&a[1]=c';
	const wasmResult = await parse(input);
	const jsResult = qsOriginal.parse(input);
	assert(deepEqual(wasmResult, jsResult), `WASM: ${JSON.stringify(wasmResult)}, JS: ${JSON.stringify(jsResult)}`);
});

// Test 4: Dot notation
await test('Parse: dot notation', async () => {
	const input = 'a.b.c=d';
	const opts = { allowDots: true };
	const wasmResult = await parse(input, opts);
	const jsResult = qsOriginal.parse(input, opts);
	assert(deepEqual(wasmResult, jsResult), `WASM: ${JSON.stringify(wasmResult)}, JS: ${JSON.stringify(jsResult)}`);
});

// Test 5: Comma separated
await test('Parse: comma separated', async () => {
	const input = 'a=1,2,3';
	const opts = { comma: true };
	const wasmResult = await parse(input, opts);
	const jsResult = qsOriginal.parse(input, opts);
	assert(deepEqual(wasmResult, jsResult), `WASM: ${JSON.stringify(wasmResult)}, JS: ${JSON.stringify(jsResult)}`);
});

// Test 6: Ignore query prefix
await test('Parse: ignoreQueryPrefix', async () => {
	const input = '?a=b&c=d';
	const opts = { ignoreQueryPrefix: true };
	const wasmResult = await parse(input, opts);
	const jsResult = qsOriginal.parse(input, opts);
	assert(deepEqual(wasmResult, jsResult), `WASM: ${JSON.stringify(wasmResult)}, JS: ${JSON.stringify(jsResult)}`);
});

// Test 7: Depth limit
await test('Parse: depth limit', async () => {
	const input = 'a[b][c][d][e]=f';
	const opts = { depth: 2 };
	const wasmResult = await parse(input, opts);
	const jsResult = qsOriginal.parse(input, opts);
	assert(deepEqual(wasmResult, jsResult), `WASM: ${JSON.stringify(wasmResult)}, JS: ${JSON.stringify(jsResult)}`);
});

// Test 8: Array limit
await test('Parse: array limit', async () => {
	const input = 'a[100]=b';
	const opts = { arrayLimit: 50 };
	const wasmResult = await parse(input, opts);
	const jsResult = qsOriginal.parse(input, opts);
	assert(deepEqual(wasmResult, jsResult), `WASM: ${JSON.stringify(wasmResult)}, JS: ${JSON.stringify(jsResult)}`);
});

// Test 9: Parameter limit
await test('Parse: parameter limit', async () => {
	const input = 'a=1&b=2&c=3&d=4&e=5';
	const opts = { parameterLimit: 3 };
	const wasmResult = await parse(input, opts);
	const jsResult = qsOriginal.parse(input, opts);
	assert(deepEqual(wasmResult, jsResult), `WASM: ${JSON.stringify(wasmResult)}, JS: ${JSON.stringify(jsResult)}`);
});

// Test 10: Strict null handling
await test('Parse: strict null handling', async () => {
	const input = 'a&b=&c';
	const opts = { strictNullHandling: true };
	const wasmResult = await parse(input, opts);
	const jsResult = qsOriginal.parse(input, opts);
	assert(deepEqual(wasmResult, jsResult), `WASM: ${JSON.stringify(wasmResult)}, JS: ${JSON.stringify(jsResult)}`);
});

// Test 11: Allow sparse arrays
await test('Parse: allow sparse', async () => {
	const input = 'a[1]=b&a[3]=c';
	const opts = { allowSparse: true };
	const wasmResult = await parse(input, opts);
	const jsResult = qsOriginal.parse(input, opts);
	assert(deepEqual(wasmResult, jsResult), `WASM: ${JSON.stringify(wasmResult)}, JS: ${JSON.stringify(jsResult)}`);
});

// Test 12: Custom delimiter
await test('Parse: custom delimiter', async () => {
	const input = 'a=1;b=2;c=3';
	const opts = { delimiter: ';' };
	const wasmResult = await parse(input, opts);
	const jsResult = qsOriginal.parse(input, opts);
	assert(deepEqual(wasmResult, jsResult), `WASM: ${JSON.stringify(wasmResult)}, JS: ${JSON.stringify(jsResult)}`);
});

// Test 13: parseArrays false
await test('Parse: parseArrays false', async () => {
	const input = 'a[0]=b&a[1]=c';
	const opts = { parseArrays: false };
	const wasmResult = await parse(input, opts);
	const jsResult = qsOriginal.parse(input, opts);
	assert(deepEqual(wasmResult, jsResult), `WASM: ${JSON.stringify(wasmResult)}, JS: ${JSON.stringify(jsResult)}`);
});

// Test 14: Unicode
await test('Parse: unicode', async () => {
	const input = 'name=%E6%97%A5%E6%9C%AC%E8%AA%9E';
	const wasmResult = await parse(input);
	const jsResult = qsOriginal.parse(input);
	assert(deepEqual(wasmResult, jsResult), `WASM: ${JSON.stringify(wasmResult)}, JS: ${JSON.stringify(jsResult)}`);
});

// Test 15: Special characters
await test('Parse: special characters', async () => {
	const input = 'msg=hello%20world%21&special=%26%3D%3F';
	const wasmResult = await parse(input);
	const jsResult = qsOriginal.parse(input);
	assert(deepEqual(wasmResult, jsResult), `WASM: ${JSON.stringify(wasmResult)}, JS: ${JSON.stringify(jsResult)}`);
});

// =============================================================================
// STRINGIFY TESTS
// =============================================================================

console.log('\n--- Stringify Tests ---\n');

// Test 16: Basic stringify
await test('Stringify: basic', async () => {
	const input = { a: 'b', c: 'd' };
	const wasmResult = await stringify(input);
	const jsResult = qsOriginal.stringify(input);
	assert(compareQueryStrings(wasmResult, jsResult), `WASM: ${wasmResult}, JS: ${jsResult}`);
});

// Test 17: Nested objects
await test('Stringify: nested objects', async () => {
	const input = { a: { b: { c: 'd' } } };
	const wasmResult = await stringify(input);
	const jsResult = qsOriginal.stringify(input);
	assert(compareQueryStrings(wasmResult, jsResult), `WASM: ${wasmResult}, JS: ${jsResult}`);
});

// Test 18: Arrays - indices format
await test('Stringify: arrays indices', async () => {
	const input = { a: ['b', 'c', 'd'] };
	const opts = { arrayFormat: 'indices' };
	const wasmResult = await stringify(input, opts);
	const jsResult = qsOriginal.stringify(input, opts);
	assert(compareQueryStrings(wasmResult, jsResult), `WASM: ${wasmResult}, JS: ${jsResult}`);
});

// Test 19: Arrays - brackets format
await test('Stringify: arrays brackets', async () => {
	const input = { a: ['b', 'c', 'd'] };
	const opts = { arrayFormat: 'brackets' };
	const wasmResult = await stringify(input, opts);
	const jsResult = qsOriginal.stringify(input, opts);
	assert(compareQueryStrings(wasmResult, jsResult), `WASM: ${wasmResult}, JS: ${jsResult}`);
});

// Test 20: Arrays - repeat format
await test('Stringify: arrays repeat', async () => {
	const input = { a: ['b', 'c', 'd'] };
	const opts = { arrayFormat: 'repeat' };
	const wasmResult = await stringify(input, opts);
	const jsResult = qsOriginal.stringify(input, opts);
	assert(compareQueryStrings(wasmResult, jsResult), `WASM: ${wasmResult}, JS: ${jsResult}`);
});

// Test 21: Arrays - comma format
await test('Stringify: arrays comma', async () => {
	const input = { a: ['b', 'c', 'd'] };
	const opts = { arrayFormat: 'comma' };
	const wasmResult = await stringify(input, opts);
	const jsResult = qsOriginal.stringify(input, opts);
	assert(compareQueryStrings(wasmResult, jsResult), `WASM: ${wasmResult}, JS: ${jsResult}`);
});

// Test 22: Dot notation
await test('Stringify: dot notation', async () => {
	const input = { a: { b: 'c' } };
	const opts = { allowDots: true, encode: false };
	const wasmResult = await stringify(input, opts);
	const jsResult = qsOriginal.stringify(input, opts);
	assert(wasmResult === jsResult, `WASM: ${wasmResult}, JS: ${jsResult}`);
});

// Test 23: Query prefix
await test('Stringify: query prefix', async () => {
	const input = { a: 'b' };
	const opts = { addQueryPrefix: true };
	const wasmResult = await stringify(input, opts);
	const jsResult = qsOriginal.stringify(input, opts);
	assert(wasmResult === jsResult, `WASM: ${wasmResult}, JS: ${jsResult}`);
});

// Test 24: Skip nulls
await test('Stringify: skip nulls', async () => {
	const input = { a: null, b: 'c', d: null };
	const opts = { skipNulls: true };
	const wasmResult = await stringify(input, opts);
	const jsResult = qsOriginal.stringify(input, opts);
	assert(compareQueryStrings(wasmResult, jsResult), `WASM: ${wasmResult}, JS: ${jsResult}`);
});

// Test 25: Strict null handling
await test('Stringify: strict null handling', async () => {
	const input = { a: null, b: 'c' };
	const opts = { strictNullHandling: true };
	const wasmResult = await stringify(input, opts);
	const jsResult = qsOriginal.stringify(input, opts);
	assert(compareQueryStrings(wasmResult, jsResult), `WASM: ${wasmResult}, JS: ${jsResult}`);
});

// Test 26: Custom delimiter
await test('Stringify: custom delimiter', async () => {
	const input = { a: '1', b: '2' };
	const opts = { delimiter: ';' };
	const wasmResult = await stringify(input, opts);
	const jsResult = qsOriginal.stringify(input, opts);
	// Compare with custom delimiter
	const wasmParts = wasmResult.split(';').sort();
	const jsParts = jsResult.split(';').sort();
	assert(deepEqual(wasmParts, jsParts), `WASM: ${wasmResult}, JS: ${jsResult}`);
});

// Test 27: Encode false
await test('Stringify: encode false', async () => {
	const input = { a: 'hello world' };
	const opts = { encode: false };
	const wasmResult = await stringify(input, opts);
	const jsResult = qsOriginal.stringify(input, opts);
	assert(wasmResult === jsResult, `WASM: ${wasmResult}, JS: ${jsResult}`);
});

// Test 28: Encode values only
await test('Stringify: encode values only', async () => {
	const input = { 'a[b]': 'c d' };
	const opts = { encodeValuesOnly: true };
	const wasmResult = await stringify(input, opts);
	const jsResult = qsOriginal.stringify(input, opts);
	assert(wasmResult === jsResult, `WASM: ${wasmResult}, JS: ${jsResult}`);
});

// Test 29: Empty arrays
await test('Stringify: empty arrays', async () => {
	const input = { a: [], b: 'c' };
	const opts = { allowEmptyArrays: true };
	const wasmResult = await stringify(input, opts);
	const jsResult = qsOriginal.stringify(input, opts);
	assert(compareQueryStrings(wasmResult, jsResult), `WASM: ${wasmResult}, JS: ${jsResult}`);
});

// Test 30: Unicode
await test('Stringify: unicode', async () => {
	const input = { name: '日本語' };
	const wasmResult = await stringify(input);
	const jsResult = qsOriginal.stringify(input);
	assert(wasmResult === jsResult, `WASM: ${wasmResult}, JS: ${jsResult}`);
});

// =============================================================================
// COMPLEX TESTS
// =============================================================================

console.log('\n--- Complex Tests ---\n');

// Test 31: Deep nested with arrays
await test('Complex: deep nested with arrays', async () => {
	const input = {
		user: {
			profile: {
				name: 'John',
				emails: ['john@example.com', 'doe@work.com'],
			},
			tags: ['admin', 'user'],
		},
	};
	const wasmResult = await stringify(input);
	const jsResult = qsOriginal.stringify(input);
	assert(compareQueryStrings(wasmResult, jsResult), `WASM: ${wasmResult}, JS: ${jsResult}`);
});

// Test 32: Round trip
await test('Complex: round trip', async () => {
	const original = { a: { b: ['c', 'd'] }, e: 'f' };
	const stringified = await stringify(original);
	const parsed = await parse(stringified);
	const jsStringified = qsOriginal.stringify(original);
	const jsParsed = qsOriginal.parse(jsStringified);
	assert(deepEqual(parsed, jsParsed), `WASM: ${JSON.stringify(parsed)}, JS: ${JSON.stringify(jsParsed)}`);
});

// Test 33: Special characters round trip
await test('Complex: special characters round trip', async () => {
	const original = {
		message: 'Hello, World! @#$%^&*()',
		unicode: '日本語テスト',
		ampersand: 'tom&jerry',
		equals: 'a=b=c',
	};
	const stringified = await stringify(original);
	const parsed = await parse(stringified);
	const jsStringified = qsOriginal.stringify(original);
	const jsParsed = qsOriginal.parse(jsStringified);
	assert(deepEqual(parsed, jsParsed), `WASM: ${JSON.stringify(parsed)}, JS: ${JSON.stringify(jsParsed)}`);
});

// Test 34: Real-world API query
await test('Complex: real-world API query', async () => {
	const input = {
		filters: {
			status: ['active', 'pending'],
			created: { $gte: '2024-01-01', $lte: '2024-12-31' },
		},
		pagination: { page: 1, limit: 25 },
		populate: ['user', 'comments'],
	};
	const wasmResult = await stringify(input);
	const jsResult = qsOriginal.stringify(input);
	assert(compareQueryStrings(wasmResult, jsResult), `WASM: ${wasmResult}, JS: ${jsResult}`);

	// Parse back
	const wasmParsed = await parse(wasmResult);
	const jsParsed = qsOriginal.parse(jsResult);
	assert(deepEqual(wasmParsed, jsParsed), `WASM: ${JSON.stringify(wasmParsed)}, JS: ${JSON.stringify(jsParsed)}`);
});

// Test 35: Boolean and number handling
await test('Complex: booleans and numbers', async () => {
	const input = { bool: true, num: 42, float: 3.14 };
	const wasmResult = await stringify(input);
	const jsResult = qsOriginal.stringify(input);
	assert(compareQueryStrings(wasmResult, jsResult), `WASM: ${wasmResult}, JS: ${jsResult}`);
});

// Test 36: Empty values
await test('Complex: empty values', async () => {
	const input = { empty: '', nested: { also: '' }, value: 'present' };
	const wasmResult = await stringify(input);
	const jsResult = qsOriginal.stringify(input);
	assert(compareQueryStrings(wasmResult, jsResult), `WASM: ${wasmResult}, JS: ${jsResult}`);
});

// Test 37: Nested arrays
await test('Complex: nested arrays', async () => {
	const input = { a: [['b', 'c'], ['d', 'e']] };
	const wasmResult = await stringify(input);
	const jsResult = qsOriginal.stringify(input);
	assert(compareQueryStrings(wasmResult, jsResult), `WASM: ${wasmResult}, JS: ${jsResult}`);

	const wasmParsed = await parse(wasmResult);
	const jsParsed = qsOriginal.parse(jsResult);
	assert(deepEqual(wasmParsed, jsParsed), `WASM: ${JSON.stringify(wasmParsed)}, JS: ${JSON.stringify(jsParsed)}`);
});

// Test 38: Combine multiple options
await test('Complex: multiple parse options', async () => {
	const input = '?a.b=1&c[0]=x&c[1]=y';
	const opts = { allowDots: true, ignoreQueryPrefix: true, depth: 5 };
	const wasmResult = await parse(input, opts);
	const jsResult = qsOriginal.parse(input, opts);
	assert(deepEqual(wasmResult, jsResult), `WASM: ${JSON.stringify(wasmResult)}, JS: ${JSON.stringify(jsResult)}`);
});

// Test 39: Combine multiple stringify options
await test('Complex: multiple stringify options', async () => {
	const input = { search: 'hello', tags: ['a', 'b'], filters: { active: true } };
	const opts = { addQueryPrefix: true, allowDots: true, arrayFormat: 'brackets' };
	const wasmResult = await stringify(input, opts);
	const jsResult = qsOriginal.stringify(input, opts);
	assert(compareQueryStrings(wasmResult, jsResult), `WASM: ${wasmResult}, JS: ${jsResult}`);
});

// Test 40: Large object
await test('Complex: large object', async () => {
	const input = {};
	for (let i = 0; i < 100; i++) {
		input[`key${i}`] = `value${i}`;
	}
	const wasmResult = await stringify(input);
	const jsResult = qsOriginal.stringify(input);
	assert(compareQueryStrings(wasmResult, jsResult), `Results differ`);

	const wasmParsed = await parse(wasmResult);
	const jsParsed = qsOriginal.parse(jsResult);
	assert(deepEqual(wasmParsed, jsParsed), `Parsed results differ`);
});

// =============================================================================
// SUMMARY
// =============================================================================


console.log('\n===========================================');
console.log(`Total: ${passed + failed} tests`);
console.log(`Passed: ${passed}`);
console.log(`Failed: ${failed}`);
console.log('===========================================\n');

if (failed > 0) {
	process.exit(1);
}
