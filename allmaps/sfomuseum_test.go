package allmaps

import (
	"testing"
)

func TestSFOMuseumDeriveImageId(t *testing.T) {

	tests := map[string]int64{
		"https://annotations.allmaps.org/maps/a0c0c652e49f4596":           1527818309,
		"https://annotations.allmaps.org/images/c157f0e8c25aa123.geojson": 1762891451,
	}

	for url, expected_id := range tests {

		id, err := DeriveSFOMuseumImageId(url)

		if err != nil {
			t.Fatalf("Failed to derive ID for '%s', %v", url, err)
		}

		if id != expected_id {
			t.Fatalf("Unexpected ID for '%s', '%d'. Expected '%d'", url, id, expected_id)
		}
	}
}
