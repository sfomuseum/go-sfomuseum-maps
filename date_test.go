package maps

import (
	"testing"
)

func TestDeriveYearLabel(t *testing.T) {

	tests := map[string][]byte{
		"193X": []byte(`{"properties":{ "edtf:inception": "193X" }}`),
		"2023": []byte(`{"properties":{ "edtf:inception": "2023", "date:inception_upper": "2023-01-31" }}`),
	}

	for expected, body := range tests {

		l, err := DeriveYearLabel(body)

		if err != nil {
			t.Fatalf("Failed to derive year label for %s, %v", expected, err)
		}

		if l != expected {
			t.Fatalf("Expected '%s' but got '%s'", expected, l)
		}
	}

}
