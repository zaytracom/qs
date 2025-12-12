const qs = require('qs');

// 10 comprehensive tests covering all qs functionality
// Each test has parse and stringify variants

const tests = [
  // Test 1: Deep nested objects with arrays, dots, and special characters
  {
    name: "deep_nested_complex",
    input: {
      user: {
        profile: {
          name: "John Doe",
          emails: ["john@example.com", "doe@work.com"],
          settings: {
            theme: "dark",
            notifications: {
              email: true,
              sms: false,
              push: ["morning", "evening"]
            }
          }
        },
        tags: ["admin", "verified", "premium"]
      },
      meta: {
        version: "2.0",
        timestamp: 1702400000
      }
    },
    parseOptions: {},
    stringifyOptions: {}
  },

  // Test 2: Sparse arrays with nulls and strict null handling
  {
    name: "sparse_arrays_nulls",
    input: {
      items: [null, "first", null, "third", null],
      config: {
        enabled: null,
        value: "test"
      },
      empty: null
    },
    parseOptions: { strictNullHandling: true, allowSparse: true },
    stringifyOptions: { strictNullHandling: true }
  },

  // Test 3: All array formats (indices, brackets, repeat, comma)
  {
    name: "array_formats_indices",
    input: {
      colors: ["red", "green", "blue"],
      numbers: [1, 2, 3, 4, 5]
    },
    parseOptions: {},
    stringifyOptions: { arrayFormat: "indices" }
  },
  {
    name: "array_formats_brackets",
    input: {
      colors: ["red", "green", "blue"],
      numbers: [1, 2, 3, 4, 5]
    },
    parseOptions: {},
    stringifyOptions: { arrayFormat: "brackets" }
  },
  {
    name: "array_formats_repeat",
    input: {
      colors: ["red", "green", "blue"],
      numbers: [1, 2, 3, 4, 5]
    },
    parseOptions: {},
    stringifyOptions: { arrayFormat: "repeat" }
  },
  {
    name: "array_formats_comma",
    input: {
      colors: ["red", "green", "blue"],
      numbers: [1, 2, 3, 4, 5]
    },
    parseOptions: { comma: true },
    stringifyOptions: { arrayFormat: "comma" }
  },

  // Test 4: Dot notation with encoded dots in keys
  {
    name: "dot_notation_encoded",
    input: {
      "user.name": "John",
      config: {
        "api.key": "secret123",
        nested: {
          "deep.value": 42
        }
      }
    },
    parseOptions: { allowDots: true, decodeDotInKeys: true },
    stringifyOptions: { allowDots: true, encodeDotInKeys: true }
  },

  // Test 5: Special characters and unicode
  {
    name: "special_chars_unicode",
    input: {
      message: "Hello, World! @#$%^&*()",
      unicode: "日本語テスト",
      emoji: "🎉🚀✨",
      spaces: "multiple   spaces   here",
      quotes: "\"quoted\" and 'single'",
      ampersand: "tom&jerry",
      equals: "a=b=c"
    },
    parseOptions: {},
    stringifyOptions: {}
  },

  // Test 6: Empty values, empty arrays, empty objects
  {
    name: "empty_values",
    input: {
      emptyString: "",
      emptyArray: [],
      nested: {
        also: {
          empty: ""
        }
      },
      normalValue: "present"
    },
    parseOptions: { allowEmptyArrays: true },
    stringifyOptions: { allowEmptyArrays: true }
  },

  // Test 7: RFC formats and charset
  {
    name: "rfc_formats",
    input: {
      space: "hello world",
      special: "a+b=c&d"
    },
    parseOptions: {},
    stringifyOptions: { format: "RFC1738" }
  },

  // Test 8: Filter, sort, and skip nulls
  {
    name: "filter_sort_skipnulls",
    input: {
      zebra: "last",
      apple: "first",
      mango: null,
      banana: "middle",
      cherry: null
    },
    parseOptions: {},
    stringifyOptions: {
      skipNulls: true,
      sort: (a, b) => a.localeCompare(b)
    }
  },

  // Test 9: Depth limits and parameter limits
  {
    name: "depth_limits",
    input: {
      a: {
        b: {
          c: {
            d: {
              e: {
                f: {
                  g: "deep"
                }
              }
            }
          }
        }
      }
    },
    parseOptions: { depth: 3 },
    stringifyOptions: {}
  },

  // Test 10: Complex real-world API query
  {
    name: "real_world_api",
    input: {
      filters: {
        status: ["active", "pending"],
        created: {
          $gte: "2024-01-01",
          $lte: "2024-12-31"
        },
        tags: {
          $in: ["important", "urgent"]
        }
      },
      pagination: {
        page: 1,
        limit: 25,
        sort: "-createdAt"
      },
      populate: ["user", "comments"],
      select: ["id", "title", "status", "createdAt"]
    },
    parseOptions: {},
    stringifyOptions: {}
  }
];

// Run tests and output results
console.log("// Auto-generated test cases from qs library\n");

for (const test of tests) {
  console.log(`// ========== ${test.name} ==========`);

  // Stringify test
  const stringified = qs.stringify(test.input, test.stringifyOptions);
  console.log(`// Stringify input: ${JSON.stringify(test.input)}`);
  console.log(`// Stringify options: ${JSON.stringify(test.stringifyOptions)}`);
  console.log(`// Stringify result: ${stringified}`);

  // Parse the stringified result back
  const parsed = qs.parse(stringified, test.parseOptions);
  console.log(`// Parse options: ${JSON.stringify(test.parseOptions)}`);
  console.log(`// Parse result: ${JSON.stringify(parsed)}`);

  console.log("");
}

// Additional edge case tests
console.log("// ========== EDGE CASES ==========\n");

// Query prefix
console.log("// Query prefix test");
const withPrefix = qs.stringify({ a: "b" }, { addQueryPrefix: true });
console.log(`// addQueryPrefix: ${withPrefix}`);
const parsedPrefix = qs.parse("?a=b", { ignoreQueryPrefix: true });
console.log(`// ignoreQueryPrefix parse: ${JSON.stringify(parsedPrefix)}`);

// Charset sentinel
console.log("\n// Charset sentinel test");
const withSentinel = qs.stringify({ a: "b" }, { charsetSentinel: true });
console.log(`// charsetSentinel: ${withSentinel}`);

// Comma round trip
console.log("\n// Comma round trip test");
const commaRT = qs.stringify({ a: ["b"] }, { arrayFormat: "comma", commaRoundTrip: true });
console.log(`// commaRoundTrip single element: ${commaRT}`);

// Encode values only
console.log("\n// Encode values only test");
const encodeValuesOnly = qs.stringify({ "a[b]": "c d" }, { encodeValuesOnly: true });
console.log(`// encodeValuesOnly: ${encodeValuesOnly}`);

// Custom delimiter
console.log("\n// Custom delimiter test");
const customDelim = qs.stringify({ a: "1", b: "2" }, { delimiter: ";" });
console.log(`// delimiter ';': ${customDelim}`);
const parsedDelim = qs.parse("a=1;b=2", { delimiter: ";" });
console.log(`// parse with ';': ${JSON.stringify(parsedDelim)}`);

// Array limit
console.log("\n// Array limit test");
const parsedArrayLimit = qs.parse("a[100]=b", { arrayLimit: 50 });
console.log(`// arrayLimit 50, a[100]=b: ${JSON.stringify(parsedArrayLimit)}`);

// Duplicates handling
console.log("\n// Duplicates handling test");
const dupCombine = qs.parse("a=1&a=2&a=3", { duplicates: "combine" });
console.log(`// duplicates combine: ${JSON.stringify(dupCombine)}`);
const dupFirst = qs.parse("a=1&a=2&a=3", { duplicates: "first" });
console.log(`// duplicates first: ${JSON.stringify(dupFirst)}`);
const dupLast = qs.parse("a=1&a=2&a=3", { duplicates: "last" });
console.log(`// duplicates last: ${JSON.stringify(dupLast)}`);

// Boolean and number handling
console.log("\n// Boolean and number handling test");
const boolNum = qs.stringify({ bool: true, num: 42, float: 3.14 });
console.log(`// booleans and numbers: ${boolNum}`);

// Nested arrays
console.log("\n// Nested arrays test");
const nestedArr = qs.stringify({ a: [["b", "c"], ["d", "e"]] });
console.log(`// nested arrays: ${nestedArr}`);
const parsedNestedArr = qs.parse(nestedArr);
console.log(`// parsed nested arrays: ${JSON.stringify(parsedNestedArr)}`);

// Very long values
console.log("\n// Long values test");
const longVal = "x".repeat(1000);
const longResult = qs.stringify({ long: longVal });
console.log(`// long value length: ${longResult.length}`);

// Many parameters
console.log("\n// Many parameters test");
const manyParams = {};
for (let i = 0; i < 100; i++) {
  manyParams[`key${i}`] = `value${i}`;
}
const manyResult = qs.stringify(manyParams);
console.log(`// 100 params stringified length: ${manyResult.length}`);
const parsedMany = qs.parse(manyResult);
console.log(`// parsed 100 params count: ${Object.keys(parsedMany).length}`);

// ISO-8859-1 charset
console.log("\n// ISO-8859-1 charset test");
const isoStr = qs.stringify({ a: "ä" }, { charset: "iso-8859-1" });
console.log(`// ISO-8859-1 'ä': ${isoStr}`);

// All options combined complex test
console.log("\n// ========== MEGA COMPLEX TEST ==========");
const megaInput = {
  users: [
    {
      id: 1,
      name: "Alice",
      roles: ["admin", "user"],
      settings: {
        theme: "dark",
        "notification.email": true
      }
    },
    {
      id: 2,
      name: "Bob",
      roles: ["user"],
      settings: {
        theme: "light",
        "notification.email": false
      }
    }
  ],
  filters: {
    active: true,
    created: {
      from: "2024-01-01",
      to: "2024-12-31"
    }
  },
  search: "hello world",
  tags: ["important", "urgent", "review"],
  pagination: {
    page: 1,
    size: 20
  },
  special: "a=b&c=d",
  unicode: "日本語",
  empty: "",
  nullVal: null
};

const megaStringified = qs.stringify(megaInput, {
  allowDots: true,
  encodeDotInKeys: true,
  arrayFormat: "indices"
});
console.log(`// Mega stringify result:`);
console.log(`// ${megaStringified}`);

const megaParsed = qs.parse(megaStringified, {
  allowDots: true,
  decodeDotInKeys: true
});
console.log(`// Mega parse result:`);
console.log(`// ${JSON.stringify(megaParsed, null, 2)}`);
