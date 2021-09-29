package iso639_3

import (
	"testing"
)

func TestFromCode(t *testing.T) {
	tests := []struct {
		code         string
		expectedName string
	}{
		{"rus", "Russian"},
		{"ru", "Russian"},
		{"de", "German"},
		{"ger", "German"},
		{"aaa", ""}, // doesn't exist
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			actual := FromCode(tt.code)

			if tt.expectedName == "" {
				if actual != nil {
					t.Errorf("FromCode() = %v, expected nil", actual)
				}
			} else if actual == nil || actual.Name != tt.expectedName {
				t.Errorf("FromCode() = %v, expected Language with english name %v", actual, tt.expectedName)
			}
		})
	}
}

func TestFromName(t *testing.T) {
	tests := []struct {
		name           string
		expectedAlpha3 string
	}{
		{"Russian", "rus"},
		{"German", "ger"},
		{"Elvish", ""}, // doesn't exist (ouch)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := FromName(tt.name)

			if tt.expectedAlpha3 == "" {
				if actual != nil {
					t.Errorf("FromCode() = %v, expected nil", actual)
				}
			} else if actual == nil || actual.ID != tt.expectedAlpha3 {
				t.Errorf("FromCode() = %v, expected Language with Alpha3 %v", actual, tt.expectedAlpha3)
			}
		})
	}
}
