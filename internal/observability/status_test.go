package observability

import "testing"

func TestDecodeBase64(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		err      bool
	}{
		{
			input:    "c3VwZXItc2VjcmV0LXBhc3N3b3JkLTEyMw==",
			expected: "super-secret-password-123",
			err:      false,
		},
		{
			input:    "c3RhY2twdWxzZQ==",
			expected: "stackpulse",
			err:      false,
		},
		{
			input:    "!!!invalidbase64!!!",
			expected: "",
			err:      true,
		},
	}

	for _, test := range tests {
		result, err := DecodeBase64(test.input)
		if test.err {
			if err == nil {
				t.Errorf("expected error decoding '%s', got nil", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error decoding '%s': %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("expected decoded value '%s', got '%s'", test.expected, result)
			}
		}
	}
}
