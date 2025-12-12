/**
 * Stress Tests for QS WASM Module
 *
 * Tests with large, deeply nested objects to ensure WASM handles
 * real-world complex data structures correctly.
 */

import { parse, stringify, createQS } from '../dist/index.js';
import qsOriginal from 'qs';
import assert from 'assert';

let passed = 0;
let failed = 0;

async function test(name, fn) {
	const start = performance.now();
	try {
		await fn();
		const elapsed = (performance.now() - start).toFixed(2);
		passed++;
		console.log(`✓ ${name} (${elapsed}ms)`);
	} catch (err) {
		const elapsed = (performance.now() - start).toFixed(2);
		failed++;
		console.error(`✗ ${name} (${elapsed}ms)`);
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

console.log('=== QS WASM Stress Tests ===\n');

const qs = await createQS();

// =============================================================================
// LARGE FLAT OBJECTS
// =============================================================================

console.log('--- Large Flat Objects ---\n');

await test('1000 keys flat object', async () => {
	const input = {};
	for (let i = 0; i < 1000; i++) {
		input[`key_${i}`] = `value_${i}_${'x'.repeat(50)}`;
	}
	const wasmResult = await stringify(input);
	const jsResult = qsOriginal.stringify(input);

	// Parse back
	const wasmParsed = await parse(wasmResult);
	const jsParsed = qsOriginal.parse(jsResult);
	assert(deepEqual(wasmParsed, jsParsed), 'Parsed results differ');
});

await test('5000 keys flat object', async () => {
	const input = {};
	for (let i = 0; i < 5000; i++) {
		input[`param${i}`] = `data${i}`;
	}
	const wasmResult = await stringify(input);
	const wasmParsed = await parse(wasmResult);
	assert(Object.keys(wasmParsed).length === 1000, 'Should respect default parameterLimit of 1000');
});

await test('10000 keys with high parameterLimit', async () => {
	const input = {};
	for (let i = 0; i < 10000; i++) {
		input[`k${i}`] = `v${i}`;
	}
	const wasmResult = await stringify(input);
	const wasmParsed = await parse(wasmResult, { parameterLimit: 20000 });
	const jsParsed = qsOriginal.parse(qsOriginal.stringify(input), { parameterLimit: 20000 });
	assert(deepEqual(wasmParsed, jsParsed), 'Parsed results differ');
});

// =============================================================================
// LARGE ARRAYS
// =============================================================================

console.log('\n--- Large Arrays ---\n');

await test('Array with 1000 elements', async () => {
	const input = { items: Array.from({ length: 1000 }, (_, i) => `item_${i}`) };
	const wasmResult = await stringify(input);
	const jsResult = qsOriginal.stringify(input);

	const wasmParsed = await parse(wasmResult);
	const jsParsed = qsOriginal.parse(jsResult);
	assert(deepEqual(wasmParsed, jsParsed), 'Parsed results differ');
});

await test('Array with 500 objects', async () => {
	const input = {
		users: Array.from({ length: 500 }, (_, i) => ({
			id: i,
			name: `User ${i}`,
			email: `user${i}@example.com`,
			active: i % 2 === 0,
		})),
	};
	const wasmResult = await stringify(input);
	const jsResult = qsOriginal.stringify(input);

	const wasmParsed = await parse(wasmResult);
	const jsParsed = qsOriginal.parse(jsResult);
	assert(deepEqual(wasmParsed, jsParsed), 'Parsed results differ');
});

await test('Multiple arrays with 200 elements each', async () => {
	const input = {
		tags: Array.from({ length: 200 }, (_, i) => `tag_${i}`),
		categories: Array.from({ length: 200 }, (_, i) => `cat_${i}`),
		labels: Array.from({ length: 200 }, (_, i) => `label_${i}`),
		flags: Array.from({ length: 200 }, (_, i) => `flag_${i}`),
		markers: Array.from({ length: 200 }, (_, i) => `marker_${i}`),
	};
	const wasmResult = await stringify(input);
	const jsResult = qsOriginal.stringify(input);

	const wasmParsed = await parse(wasmResult);
	const jsParsed = qsOriginal.parse(jsResult);
	assert(deepEqual(wasmParsed, jsParsed), 'Parsed results differ');
});

// =============================================================================
// DEEPLY NESTED OBJECTS
// =============================================================================

console.log('\n--- Deeply Nested Objects ---\n');

await test('5 levels deep nesting', async () => {
	const input = {
		level1: {
			level2: {
				level3: {
					level4: {
						level5: {
							value: 'deep_value',
							array: ['a', 'b', 'c'],
						},
					},
				},
			},
		},
	};
	const wasmResult = await stringify(input);
	const jsResult = qsOriginal.stringify(input);

	const wasmParsed = await parse(wasmResult);
	const jsParsed = qsOriginal.parse(jsResult);
	assert(deepEqual(wasmParsed, jsParsed), 'Parsed results differ');
});

await test('Complex nested structure with 100 branches', async () => {
	const input = { root: {} };
	for (let i = 0; i < 100; i++) {
		input.root[`branch_${i}`] = {
			id: i,
			data: {
				name: `Branch ${i}`,
				values: [i * 1, i * 2, i * 3],
				meta: {
					created: `2024-01-${String(i % 28 + 1).padStart(2, '0')}`,
					tags: [`tag_${i}_a`, `tag_${i}_b`],
				},
			},
		};
	}
	const wasmResult = await stringify(input);
	const jsResult = qsOriginal.stringify(input);

	const wasmParsed = await parse(wasmResult);
	const jsParsed = qsOriginal.parse(jsResult);
	assert(deepEqual(wasmParsed, jsParsed), 'Parsed results differ');
});

await test('Tree structure with many leaves', async () => {
	const buildTree = (depth, breadth) => {
		if (depth === 0) return `leaf_${Math.random().toString(36).slice(2, 8)}`;
		const node = {};
		for (let i = 0; i < breadth; i++) {
			node[`child_${i}`] = buildTree(depth - 1, breadth);
		}
		return node;
	};
	// 4 levels deep, 5 children each = 5^4 = 625 leaves
	const input = { tree: buildTree(4, 5) };

	const wasmResult = await stringify(input);
	const jsResult = qsOriginal.stringify(input);

	const wasmParsed = await parse(wasmResult);
	const jsParsed = qsOriginal.parse(jsResult);
	assert(deepEqual(wasmParsed, jsParsed), 'Parsed results differ');
});

// =============================================================================
// REAL-WORLD COMPLEX STRUCTURES
// =============================================================================

console.log('\n--- Real-World Complex Structures ---\n');

await test('E-commerce product catalog (50 products)', async () => {
	const input = {
		products: Array.from({ length: 50 }, (_, i) => ({
			id: `prod_${i}`,
			sku: `SKU-${100000 + i}`,
			name: `Product ${i} with a very long descriptive name`,
			description: `This is a detailed description for product ${i}. `.repeat(5),
			price: {
				amount: (i + 1) * 9.99,
				currency: 'USD',
				discount: i % 3 === 0 ? { percent: 10, code: `SAVE${i}` } : null,
			},
			inventory: {
				quantity: i * 10,
				warehouses: [
					{ id: 'WH1', stock: i * 5 },
					{ id: 'WH2', stock: i * 5 },
				],
			},
			categories: [`cat_${i % 10}`, `cat_${(i + 5) % 10}`],
			tags: Array.from({ length: 5 }, (_, j) => `tag_${i}_${j}`),
			attributes: {
				color: ['red', 'blue', 'green'][i % 3],
				size: ['S', 'M', 'L', 'XL'][i % 4],
				weight: `${(i + 1) * 0.5}kg`,
			},
			metadata: {
				created: `2024-${String((i % 12) + 1).padStart(2, '0')}-15`,
				updated: `2024-12-${String((i % 28) + 1).padStart(2, '0')}`,
				version: i + 1,
			},
		})),
		pagination: {
			page: 1,
			limit: 50,
			total: 1000,
			hasNext: true,
			hasPrev: false,
		},
		filters: {
			applied: {
				priceRange: { min: 0, max: 500 },
				categories: ['cat_1', 'cat_2', 'cat_3'],
				inStock: true,
			},
			available: {
				colors: ['red', 'blue', 'green', 'yellow', 'black', 'white'],
				sizes: ['XS', 'S', 'M', 'L', 'XL', 'XXL'],
			},
		},
		sort: { field: 'price', order: 'asc' },
	};

	const wasmResult = await stringify(input);
	const jsResult = qsOriginal.stringify(input);

	// Parse with high limit to get all data
	const wasmParsed = await parse(wasmResult, { parameterLimit: 10000 });
	const jsParsed = qsOriginal.parse(jsResult, { parameterLimit: 10000 });
	assert(deepEqual(wasmParsed, jsParsed), 'Parsed results differ');
});

await test('Analytics event batch (100 events)', async () => {
	const input = {
		batch: {
			id: 'batch_12345',
			timestamp: '2024-12-12T10:30:00Z',
			source: 'web_client',
		},
		events: Array.from({ length: 100 }, (_, i) => ({
			id: `evt_${i}`,
			type: ['pageview', 'click', 'scroll', 'form_submit'][i % 4],
			timestamp: `2024-12-12T10:${String(i % 60).padStart(2, '0')}:00Z`,
			user: {
				id: `user_${i % 20}`,
				session: `sess_${i % 10}`,
				properties: {
					browser: ['Chrome', 'Firefox', 'Safari'][i % 3],
					os: ['Windows', 'macOS', 'Linux'][i % 3],
					device: ['desktop', 'mobile', 'tablet'][i % 3],
				},
			},
			page: {
				url: `https://example.com/page/${i}`,
				title: `Page ${i} Title`,
				referrer: i > 0 ? `https://example.com/page/${i - 1}` : null,
			},
			data: {
				element: i % 4 === 1 ? { id: `btn_${i}`, class: 'cta-button' } : null,
				scroll_depth: i % 4 === 2 ? i * 10 : null,
				form_fields: i % 4 === 3 ? ['name', 'email', 'message'] : null,
			},
		})),
		context: {
			app: { name: 'MyApp', version: '2.0.0', build: '1234' },
			campaign: { source: 'google', medium: 'cpc', name: 'winter_sale' },
		},
	};

	const wasmResult = await stringify(input, { skipNulls: true });
	const jsResult = qsOriginal.stringify(input, { skipNulls: true });

	// Parse with high limit to get all data
	const wasmParsed = await parse(wasmResult, { parameterLimit: 20000 });
	const jsParsed = qsOriginal.parse(jsResult, { parameterLimit: 20000 });
	assert(deepEqual(wasmParsed, jsParsed), 'Parsed results differ');
});

await test('GraphQL-style query with nested selections', async () => {
	const input = {
		query: {
			users: {
				fields: ['id', 'name', 'email', 'createdAt'],
				where: {
					AND: [
						{ status: { eq: 'active' } },
						{ role: { in: ['admin', 'editor', 'viewer'] } },
						{
							OR: [
								{ department: { eq: 'engineering' } },
								{ department: { eq: 'product' } },
							],
						},
					],
				},
				orderBy: [
					{ field: 'createdAt', direction: 'desc' },
					{ field: 'name', direction: 'asc' },
				],
				pagination: { first: 20, after: 'cursor_abc123' },
				include: {
					posts: {
						fields: ['id', 'title', 'status'],
						where: { published: { eq: true } },
						limit: 5,
					},
					comments: {
						fields: ['id', 'content', 'createdAt'],
						orderBy: { field: 'createdAt', direction: 'desc' },
						limit: 10,
					},
					profile: {
						fields: ['avatar', 'bio', 'social'],
					},
				},
			},
		},
		variables: {
			userId: 'user_123',
			includeDeleted: false,
			locale: 'en-US',
		},
	};

	const wasmResult = await stringify(input);
	const jsResult = qsOriginal.stringify(input);

	const wasmParsed = await parse(wasmResult);
	const jsParsed = qsOriginal.parse(jsResult);
	assert(deepEqual(wasmParsed, jsParsed), 'Parsed results differ');
});

await test('Multi-tenant SaaS configuration', async () => {
	const input = {
		tenants: Array.from({ length: 20 }, (_, i) => ({
			id: `tenant_${i}`,
			name: `Organization ${i}`,
			plan: ['free', 'starter', 'pro', 'enterprise'][i % 4],
			settings: {
				features: {
					analytics: i % 4 >= 1,
					api_access: i % 4 >= 2,
					sso: i % 4 >= 3,
					custom_domain: i % 4 >= 3,
					white_label: i % 4 === 3,
				},
				limits: {
					users: [5, 25, 100, 'unlimited'][i % 4],
					storage_gb: [1, 10, 100, 1000][i % 4],
					api_calls_per_month: [1000, 10000, 100000, 'unlimited'][i % 4],
				},
				branding: {
					logo_url: `https://cdn.example.com/tenants/${i}/logo.png`,
					primary_color: `#${((i * 123456) % 0xffffff).toString(16).padStart(6, '0')}`,
					custom_css: i % 4 === 3 ? '.header { background: custom; }' : null,
				},
				integrations: Array.from({ length: i % 5 + 1 }, (_, j) => ({
					type: ['slack', 'github', 'jira', 'salesforce', 'hubspot'][j % 5],
					enabled: true,
					config: { webhook_url: `https://hooks.example.com/${i}/${j}` },
				})),
			},
			users: Array.from({ length: Math.min(i + 1, 10) }, (_, j) => ({
				id: `user_${i}_${j}`,
				role: j === 0 ? 'owner' : ['admin', 'member'][j % 2],
				permissions: j === 0 ? ['all'] : ['read', 'write'].slice(0, j % 2 + 1),
			})),
		})),
		global: {
			maintenance_mode: false,
			announcement: {
				active: true,
				message: 'System upgrade scheduled for this weekend',
				type: 'info',
			},
			feature_flags: {
				new_dashboard: { enabled: true, rollout_percent: 50 },
				ai_assistant: { enabled: false, waitlist: true },
				mobile_app: { enabled: true, platforms: ['ios', 'android'] },
			},
		},
	};

	const wasmResult = await stringify(input, { skipNulls: true });
	const jsResult = qsOriginal.stringify(input, { skipNulls: true });

	const wasmParsed = await parse(wasmResult);
	const jsParsed = qsOriginal.parse(jsResult);
	assert(deepEqual(wasmParsed, jsParsed), 'Parsed results differ');
});

// =============================================================================
// EDGE CASES WITH LARGE DATA
// =============================================================================

console.log('\n--- Edge Cases with Large Data ---\n');

await test('Long string values (1000 chars each)', async () => {
	const input = {};
	for (let i = 0; i < 100; i++) {
		input[`field_${i}`] = 'x'.repeat(1000) + `_${i}`;
	}
	const wasmResult = await stringify(input);
	const jsResult = qsOriginal.stringify(input);

	const wasmParsed = await parse(wasmResult, { parameterLimit: 10000 });
	const jsParsed = qsOriginal.parse(jsResult, { parameterLimit: 10000 });
	assert(deepEqual(wasmParsed, jsParsed), 'Parsed results differ');
});

await test('Unicode strings in large object', async () => {
	const unicodeStrings = [
		'日本語テスト',
		'中文测试',
		'한국어 테스트',
		'Тест на русском',
		'اختبار عربي',
		'🎉🚀✨💻🌍',
		'Ñoño español',
		'Ελληνικά',
		'עברית',
		'ไทย',
	];
	const input = {};
	for (let i = 0; i < 200; i++) {
		input[`field_${i}`] = unicodeStrings[i % unicodeStrings.length] + `_${i}`;
	}
	const wasmResult = await stringify(input);
	const jsResult = qsOriginal.stringify(input);

	const wasmParsed = await parse(wasmResult);
	const jsParsed = qsOriginal.parse(jsResult);
	assert(deepEqual(wasmParsed, jsParsed), 'Parsed results differ');
});

await test('Special characters stress test', async () => {
	const specials = ['&', '=', '?', '#', '%', '+', ' ', '\t', '\n', '/', '\\', '"', "'", '<', '>'];
	const input = {};
	for (let i = 0; i < 100; i++) {
		const chars = specials.map((c) => c.repeat(i % 5 + 1)).join('');
		input[`key_${i}`] = `value_${chars}_end`;
	}
	const wasmResult = await stringify(input);
	const jsResult = qsOriginal.stringify(input);

	const wasmParsed = await parse(wasmResult);
	const jsParsed = qsOriginal.parse(jsResult);
	assert(deepEqual(wasmParsed, jsParsed), 'Parsed results differ');
});

await test('Sparse arrays (indexes up to 1000)', async () => {
	const input = { sparse: [] };
	for (let i = 0; i < 50; i++) {
		input.sparse[i * 20] = `value_${i}`;
	}
	const wasmResult = await stringify(input);
	const jsResult = qsOriginal.stringify(input);

	// Parse with high array limit
	const wasmParsed = await parse(wasmResult, { allowSparse: true, arrayLimit: 1000 });
	const jsParsed = qsOriginal.parse(jsResult, { allowSparse: true, arrayLimit: 1000 });
	assert(deepEqual(wasmParsed, jsParsed), 'Parsed results differ');
});

await test('Mixed types in large structure', async () => {
	const input = {
		strings: Array.from({ length: 100 }, (_, i) => `str_${i}`),
		numbers: Array.from({ length: 100 }, (_, i) => i * 1.5),
		booleans: Array.from({ length: 100 }, (_, i) => i % 2 === 0),
		nested: Array.from({ length: 50 }, (_, i) => ({
			id: i,
			name: `Item ${i}`,
			active: i % 3 === 0,
			score: i * 0.1,
			tags: [`a_${i}`, `b_${i}`],
		})),
	};
	const wasmResult = await stringify(input);
	const jsResult = qsOriginal.stringify(input);

	const wasmParsed = await parse(wasmResult);
	const jsParsed = qsOriginal.parse(jsResult);
	assert(deepEqual(wasmParsed, jsParsed), 'Parsed results differ');
});

// =============================================================================
// PERFORMANCE COMPARISON
// =============================================================================

console.log('\n--- Performance Comparison ---\n');

await test('Perf: stringify 1000 key object', async () => {
	const input = {};
	for (let i = 0; i < 1000; i++) {
		input[`key${i}`] = `value${i}`;
	}

	const wasmStart = performance.now();
	for (let i = 0; i < 10; i++) {
		await stringify(input);
	}
	const wasmTime = performance.now() - wasmStart;

	const jsStart = performance.now();
	for (let i = 0; i < 10; i++) {
		qsOriginal.stringify(input);
	}
	const jsTime = performance.now() - jsStart;

	console.log(`    WASM: ${wasmTime.toFixed(2)}ms, JS: ${jsTime.toFixed(2)}ms (10 iterations)`);
	// Just ensure it completes, not asserting performance
});

await test('Perf: parse 1000 params', async () => {
	const params = Array.from({ length: 1000 }, (_, i) => `key${i}=value${i}`).join('&');

	const wasmStart = performance.now();
	for (let i = 0; i < 10; i++) {
		await parse(params, { parameterLimit: 2000 });
	}
	const wasmTime = performance.now() - wasmStart;

	const jsStart = performance.now();
	for (let i = 0; i < 10; i++) {
		qsOriginal.parse(params, { parameterLimit: 2000 });
	}
	const jsTime = performance.now() - jsStart;

	console.log(`    WASM: ${wasmTime.toFixed(2)}ms, JS: ${jsTime.toFixed(2)}ms (10 iterations)`);
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
