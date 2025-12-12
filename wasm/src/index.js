/**
 * QS WASM Module for Node.js
 */

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import { Go } from './wasm-exec.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

let instance = null;
let go = null;

/**
 * Initialize the WASM module
 * @returns {Promise<void>}
 */
async function init() {
	if (instance) return;

	go = new Go();
	const wasmPath = path.join(__dirname, 'qs.wasm');
	const wasmBuffer = fs.readFileSync(wasmPath);
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
 * Create a sync QS instance (must call init first)
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

export default { parse, stringify, createQS };
