package iso639_3

// LanguageScope represents language scope as defined in ISO 639-3
type LanguageScope rune

// LanguageScope represents language scope as defined in ISO 639-3
type LanguageType rune

const (
	LanguageTypeIndividual    LanguageScope = 'I'
	LanguageTypeSpecial       LanguageScope = 'S'
	LanguageTypeMacrolanguage LanguageScope = 'M'

	LanguageScopeLiving      LanguageType = 'L'
	LanguageScopeHistorical  LanguageType = 'H'
	LanguageScopeAncient     LanguageType = 'A'
	LanguageScopeExtinct     LanguageType = 'E'
	LanguageScopeConstructed LanguageType = 'C'
	LanguageScopeSpecial     LanguageType = 'S'
)

// Language holds language info - all ISO 639 codes along with name and some additional info
type Language struct {
	Part3        string // ISO639-3 code
	Part2B       string // ISO639-2 bibliographic code
	Part2T       string // ISO639-2 terminology code
	Part1        string // ISO639-1 code
	Scope        LanguageScope
	LanguageType LanguageType
	Name         string
	Comment      string
}

//go:generate go run cmd/generator.go -o lang-db.go

// FromPart3Code looks up language for given ISO639-3 three-symbol code.
// Returns nil if not found
func FromPart3Code(code string) *Language {
	if l, ok := LanguagesPart3[code]; ok {
		return &l
	}
	return nil
}

// FromPart2Code looks up language for given ISO639-2 (both bibliographic or terminology) three-symbol code.
// Returns nil if not found
func FromPart2Code(code string) *Language {
	if l, ok := LanguagesPart2[code]; ok {
		return &l
	}
	return nil
}

// FromPart1Code looks up language for given ISO639-1 two-symbol code.
// Returns nil if not found
func FromPart1Code(code string) *Language {
	if l, ok := LanguagesPart1[code]; ok {
		return &l
	}
	return nil
}

// FromAnyCode looks up language for given code.
// For three-symbol codes it tries ISO639-3 first, then ISO639-2.
// For two-symbol codes it tries ISO639-1.
// Returns nil if not found
func FromAnyCode(code string) *Language {
	codeLen := len(code)

	if codeLen == 3 {
		ret := FromPart3Code(code)
		if ret == nil {
			ret = FromPart2Code(code)
		}
		return ret
	}

	if codeLen == 2 {
		return FromPart1Code(code)
	}

	return nil
}

// FromName looks up language for given reference name.
// Returns nil if not found
func FromName(name string) *Language {
	for _, l := range LanguagesPart3 {
		if l.Name == name {
			return &l
		}
	}
	return nil
}
