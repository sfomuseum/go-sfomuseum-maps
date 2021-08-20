package qgis

import (
	"context"
	"os"
	"testing"
)

func TestUnmarshalGroundControlPoints(t *testing.T) {

	ctx := context.Background()

	path := "../fixtures/1930.points"
	fh, err := os.Open(path)

	if err != nil {
		t.Fatalf("Failed to open '%s', %v", path, err)
	}

	defer fh.Close()

	gcp, err := UnmarshalGroundControlPoints(ctx, fh)

	if err != nil {
		t.Fatalf("Failed to unmarshal ground control points, %v", err)
	}

	expected_csr := `#CRS: GEOGCRS["WGS 84",DATUM["World Geodetic System 1984",ELLIPSOID["WGS 84",6378137,298.257223563,LENGTHUNIT["metre",1]]],PRIMEM["Greenwich",0,ANGLEUNIT["degree",0.0174532925199433]],CS[ellipsoidal,2],AXIS["geodetic latitude (Lat)",north,ORDER[1],ANGLEUNIT["degree",0.0174532925199433]],AXIS["geodetic longitude (Lon)",east,ORDER[2],ANGLEUNIT["degree",0.0174532925199433]],USAGE[SCOPE["unknown"],AREA["World"],BBOX[-90,-180,90,180]],ID["EPSG",4326]]`

	if gcp.CSR != expected_csr {
		t.Fatalf("Unexpected CSR '%s'", gcp.CSR)
	}
}
