/**
 * QS WASM Module for Browser
 */

import { Go } from './wasm-exec.js';

let instance = null;
let go = null;
let wasmUrl = null;

/**
 * Set custom WASM URL (call before using parse/stringify)
 * @param {string} url - URL to qs.wasm file
 */
export function setWasmUrl(url) {
	wasmUrl = url;
}

/**
 * Initialize the WASM module
 * @returns {Promise<void>}
 */
async function init() {
	if (instance) return;

	go = new Go();

	// Determine WASM URL
	let url = wasmUrl;
	if (!url) {
		// Try to find wasm relative to this script
		if (typeof import.meta.url !== 'undefined') {
			url = new URL('./qs.wasm', import.meta.url).href;
		} else {
			url = '/qs.wasm';
		}
	}

	const response = await fetch(url);
	if (!response.ok) {
		throw new Error(`Failed to load WASM: ${response.status} ${response.statusText}`);
	}

	const wasmBuffer = await response.arrayBuffer();
	const result = await WebAssembly.instantiate(wasmBuffer, go.importObject);
	instance = result.instance;
	go.run(instance);
}

/**
 * Parse a query string into an object
 * @param {string} queryString - The query string to parse
 * @param {Object} [options] - Parse options
 * @returns {Promise<Object>} Parsed object
 */
export async function parse(queryString, options = {}) {
	await init();
	const result = globalThis.qsParse(queryString, options);
	if (result.error) {
		throw new Error(result.error);
	}
	return JSON.parse(result.result);
}

/**
 * Stringify an object to a query string
 * @param {Object} obj - The object to stringify
 * @param {Object} [options] - Stringify options
 * @returns {Promise<string>} Query string
 */
export async function stringify(obj, options = {}) {
	await init();
	const result = globalThis.qsStringify(JSON.stringify(obj), options);
	if (typeof result === 'object' && result.error) {
		throw new Error(result.error);
	}
	return result;
}

/**
 * Create a sync QS instance (must call after init)
 * @returns {Promise<{parse: Function, stringify: Function}>}
 */
export async function createQS() {
	await init();
	return {
		parse(queryString, options = {}) {
			const result = globalThis.qsParse(queryString, options);
			if (result.error) {
				throw new Error(result.error);
			}
			return JSON.parse(result.result);
		},
		stringify(obj, options = {}) {
			const result = globalThis.qsStringify(JSON.stringify(obj), options);
			if (typeof result === 'object' && result.error) {
				throw new Error(result.error);
			}
			return result;
		}
	};
}

export default { parse, stringify, createQS, setWasmUrl };
