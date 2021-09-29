package iso639_3

type Language struct {
	ID           string
	Part2B       string
	Part2T       string
	Part1        string
	Scope        string
	LanguageType string
	Name         string
	Comment      string
}

//go:generate go run cmd/generator.go -o lang-db.go

// FromCode looks up language for given ISO639-3 code.
// Returns nil if not found
func FromCode(code string) *Language {
	if l, ok := Languages[code]; ok {
		return &l
	}
	return nil
}

// FromName looks up language for given reference name.
// Returns nil if not found
func FromName(name string) *Language {
	for _, l := range Languages {
		if l.Name == name {
			return &l
		}
	}
	return nil
}
