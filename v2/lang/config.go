package lang

// Charset is the interpreted charset for decoding (selected by config/sentinel).
type Charset uint8

const (
	CharsetUTF8 Charset = iota
	CharsetISO88591
)

// Profile controls strictness/performance trade-offs.
type Profile uint8

const (
	ProfileFast Profile = iota
	ProfileStrict
)

// Flags control parsing behavior via bitmask.
type Flags uint32

const (
	FlagAllowDots Flags = 1 << iota
	FlagAllowEmptyArrays
	FlagAllowPrototypes
	FlagAllowSparse
	FlagComma
	FlagStrictNullHandling
	FlagStrictDepth
	FlagIgnoreQueryPrefix
	FlagCharsetSentinel
	FlagInterpretNumericEntities
	FlagDecodeDotInKeys
	FlagThrowOnLimitExceeded
)

func (f Flags) Has(flag Flags) bool { return f&flag != 0 }

const (
	DefaultDepth          = 5
	DefaultArrayLimit     = 20
	DefaultParameterLimit = 1000
	DefaultDelimiter      = byte('&')
)

// Config controls the AST parser behavior.
type Config struct {
	Flags Flags

	// ProfileFast skips validation; ProfileStrict validates percent encoding.
	Profile Profile

	Delimiter byte

	Depth          uint16
	ArrayLimit     uint16
	ParameterLimit uint16

	ParseArrays  bool
	PlainObjects bool

	Charset Charset
}

// DefaultConfig returns a config matching the language defaults.
func DefaultConfig() Config {
	return Config{
		Flags:          0,
		Profile:        ProfileFast,
		Delimiter:      DefaultDelimiter,
		Depth:          DefaultDepth,
		ArrayLimit:     DefaultArrayLimit,
		ParameterLimit: DefaultParameterLimit,
		ParseArrays:    true,
		PlainObjects:   false,
		Charset:        CharsetUTF8,
	}
}
