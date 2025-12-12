import { parse, stringify } from '../dist/index.js';
import qs from 'qs';

// Test sortArrayIndices option
console.log('=== Testing sortArrayIndices ===\n');

const testArr = { arr: ['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l'] };

// WASM with sortArrayIndices: true should match JS qs with sort
const wasmSorted = await stringify(testArr, { sortArrayIndices: true });
const jsSorted = qs.stringify(testArr, { sort: (a, b) => a.localeCompare(b) });

console.log('WASM (sortArrayIndices: true):');
console.log(wasmSorted.substring(0, 120) + '...\n');

console.log('JS qs (sort: localeCompare):');
console.log(jsSorted.substring(0, 120) + '...\n');

console.log('Results match:', wasmSorted === jsSorted);

if (wasmSorted !== jsSorted) {
  for (let i = 0; i < Math.min(wasmSorted.length, jsSorted.length); i++) {
    if (wasmSorted[i] !== jsSorted[i]) {
      console.log(`First diff at position ${i}:`);
      console.log('WASM:', wasmSorted.slice(Math.max(0, i-20), i+30));
      console.log('JS:  ', jsSorted.slice(Math.max(0, i-20), i+30));
      break;
    }
  }
  process.exit(1);
}

console.log('\n✓ sortArrayIndices works correctly!');
