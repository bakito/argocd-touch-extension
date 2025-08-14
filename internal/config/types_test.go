package config

import (
	"testing"
)

func TestResources_validateKeys(t *testing.T) {
	tests := []struct {
		name        string
		resources   Resources
		expectError bool
	}{
		{
			name:        "valid keys",
			resources:   Resources{"validKey1": Resource{}, "another_valid_key": Resource{}},
			expectError: false,
		},
		{
			name:        "invalid key '-'' character",
			resources:   Resources{"another_valid-key": Resource{}},
			expectError: true,
		},
		{
			name:        "invalid key special character",
			resources:   Resources{"invalid key!": Resource{}},
			expectError: true,
		},
		{
			name:        "invalid key empty string",
			resources:   Resources{"": Resource{}},
			expectError: true,
		},
		{
			name:        "mixed valid and invalid keys",
			resources:   Resources{"validKey": Resource{}, "invalid$key": Resource{}},
			expectError: true,
		},
		{
			name:        "no resources",
			resources:   Resources{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.resources.validateKeys()
			if (err != nil) != tt.expectError {
				t.Errorf("validateKeys() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}
