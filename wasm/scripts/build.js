#!/usr/bin/env node

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const root = path.join(__dirname, '..');
const src = path.join(root, 'src');
const dist = path.join(root, 'dist');

// Ensure dist exists
if (!fs.existsSync(dist)) {
	fs.mkdirSync(dist, { recursive: true });
}

// Copy files from src to dist
const files = ['index.js', 'browser.js', 'wasm-exec.js', 'index.d.ts'];

for (const file of files) {
	const srcPath = path.join(src, file);
	const distPath = path.join(dist, file);

	if (fs.existsSync(srcPath)) {
		// Update imports to be relative to dist
		let content = fs.readFileSync(srcPath, 'utf8');

		// Fix wasm-exec.js import path
		content = content.replace(/from ['"]\.\/wasm-exec\.js['"]/g, "from './wasm-exec.js'");

		fs.writeFileSync(distPath, content);
		console.log(`Copied: ${file}`);
	}
}

// Copy qs.wasm if it exists (should be built first)
const wasmSrc = path.join(root, 'qs.wasm');
const wasmDist = path.join(dist, 'qs.wasm');
if (fs.existsSync(wasmSrc)) {
	fs.copyFileSync(wasmSrc, wasmDist);
	console.log('Copied: qs.wasm');
} else {
	console.warn('Warning: qs.wasm not found. Run "npm run build:wasm" first.');
}

console.log('Build complete!');
