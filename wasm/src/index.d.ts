export interface ParseOptions {
	/** Enable dot notation (a.b → {a:{b:...}}) */
	allowDots?: boolean;
	/** Allow key[] with empty value to be [] */
	allowEmptyArrays?: boolean;
	/** Preserve sparse arrays */
	allowSparse?: boolean;
	/** Max index for array parsing (default: 20) */
	arrayLimit?: number;
	/** Parse comma-separated values as arrays */
	comma?: boolean;
	/** Max nesting depth (default: 5) */
	depth?: number;
	/** Strip leading ? */
	ignoreQueryPrefix?: boolean;
	/** Max parameters to parse (default: 1000) */
	parameterLimit?: number;
	/** Enable array parsing (default: true) */
	parseArrays?: boolean;
	/** Treat key (no value) as null */
	strictNullHandling?: boolean;
	/** Parameter delimiter (default: &) */
	delimiter?: string;
}

export interface StringifyOptions {
	/** Add leading ? */
	addQueryPrefix?: boolean;
	/** Use dot notation */
	allowDots?: boolean;
	/** Output key[] for empty arrays */
	allowEmptyArrays?: boolean;
	/** Array format: indices, brackets, repeat, or comma */
	arrayFormat?: 'indices' | 'brackets' | 'repeat' | 'comma';
	/** URL encode output (default: true) */
	encode?: boolean;
	/** Only encode values, not keys */
	encodeValuesOnly?: boolean;
	/** Skip null/undefined values */
	skipNulls?: boolean;
	/** Omit = for null values */
	strictNullHandling?: boolean;
	/** Parameter delimiter (default: &) */
	delimiter?: string;
}

export interface QS {
	parse(queryString: string, options?: ParseOptions): object;
	stringify(obj: object, options?: StringifyOptions): string;
}

/**
 * Parse a query string into an object
 */
export function parse(queryString: string, options?: ParseOptions): Promise<object>;

/**
 * Stringify an object to a query string
 */
export function stringify(obj: object, options?: StringifyOptions): Promise<string>;

/**
 * Create a sync QS instance
 */
export function createQS(): Promise<QS>;

declare const qs: {
	parse: typeof parse;
	stringify: typeof stringify;
	createQS: typeof createQS;
};

export default qs;
