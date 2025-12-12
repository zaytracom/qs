import { parse, stringify, createQS } from '../dist/index.js';

async function main() {
	console.log('=== QS WASM Test ===\n');

	// Test async API
	console.log('1. Async parse:');
	const result1 = await parse('a=b&c=d');
	console.log('   Input: "a=b&c=d"');
	console.log('   Output:', result1);

	console.log('\n2. Async stringify:');
	const result2 = await stringify({ a: 'b', c: 'd' });
	console.log('   Input: { a: "b", c: "d" }');
	console.log('   Output:', result2);

	// Test sync API
	console.log('\n3. Sync API (createQS):');
	const qs = await createQS();

	const parsed = qs.parse('a[b][c]=d');
	console.log('   parse("a[b][c]=d"):', parsed);

	const str = qs.stringify({ a: { b: 'c' } }, { encode: false });
	console.log('   stringify({ a: { b: "c" } }):', str);

	// Test options
	console.log('\n4. With options:');
	const dotParsed = await parse('a.b.c=d', { allowDots: true });
	console.log('   parse("a.b.c=d", { allowDots: true }):', dotParsed);

	const commaParsed = await parse('a=1,2,3', { comma: true });
	console.log('   parse("a=1,2,3", { comma: true }):', commaParsed);

	const bracketsStr = await stringify({ a: ['b', 'c'] }, { arrayFormat: 'brackets', encode: false });
	console.log('   stringify({ a: ["b", "c"] }, { arrayFormat: "brackets" }):', bracketsStr);

	console.log('\n=== All tests passed! ===');
}

main().catch(console.error);
