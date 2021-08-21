package qgis

import (
	"bufio"
	"bytes"
	"context"
	"io"
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

	expected_csr := `GEOGCRS["WGS 84",DATUM["World Geodetic System 1984",ELLIPSOID["WGS 84",6378137,298.257223563,LENGTHUNIT["metre",1]]],PRIMEM["Greenwich",0,ANGLEUNIT["degree",0.0174532925199433]],CS[ellipsoidal,2],AXIS["geodetic latitude (Lat)",north,ORDER[1],ANGLEUNIT["degree",0.0174532925199433]],AXIS["geodetic longitude (Lon)",east,ORDER[2],ANGLEUNIT["degree",0.0174532925199433]],USAGE[SCOPE["unknown"],AREA["World"],BBOX[-90,-180,90,180]],ID["EPSG",4326]]`

	if gcp.CRS != expected_csr {
		t.Fatalf("Unexpected CRS '%s'", gcp.CRS)
	}
}

func TestMarshalGroundControlPoints(t *testing.T) {

	ctx := context.Background()

	path := "../fixtures/1930.points"
	fh, err := os.Open(path)

	if err != nil {
		t.Fatalf("Failed to open '%s', %v", path, err)
	}

	defer fh.Close()

	body, err := io.ReadAll(fh)

	if err != nil {
		t.Fatalf("Failed to read '%s', %v", path, err)
	}

	br := bytes.NewReader(body)
	gcp, err := UnmarshalGroundControlPoints(ctx, br)

	var buf bytes.Buffer
	wr := bufio.NewWriter(&buf)

	err = Marshal(ctx, gcp, wr)

	if err != nil {
		t.Fatalf("Failed to marshal GCP, %v", err)
	}

	wr.Flush()

	new_body := buf.Bytes()

	if bytes.Compare(body, new_body) != 0 {
		t.Fatalf("Marshal doesn't match: '%s'\n'%s'", string(body), string(new_body))
	}

}
