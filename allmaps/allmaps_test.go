package allmaps

import (
	"testing"
)

func TestDeriveImageId(t *testing.T) {

	tests := map[string]string{
		"https://annotations.allmaps.org/maps/a0c0c652e49f4596":           "ced8faec8c108002",
		"https://annotations.allmaps.org/images/c157f0e8c25aa123.geojson": "c157f0e8c25aa123",
	}

	for url, expected_id := range tests {

		id, err := DeriveImageId(url)

		if err != nil {
			t.Fatalf("Failed to derive ID for '%s', %v", url, err)
		}

		if id != expected_id {
			t.Fatalf("Unexpected ID for '%s', '%s'. Expected '%s'", url, id, expected_id)
		}
	}
}
