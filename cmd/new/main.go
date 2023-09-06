// new is a command-line tool for making a new GeoJSON record for a map with a default geometry
// (the bounding box for SFO) to be edited in QGIS (or equivalent)
package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log"

	"github.com/paulmach/orb/geojson"
	"github.com/sfomuseum/go-flags/multi"
	sfom_writer "github.com/sfomuseum/go-sfomuseum-writer/v3"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-export/v2"
	"github.com/whosonfirst/go-whosonfirst-feature/geometry"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader"
	"github.com/whosonfirst/go-writer/v3"
)

//go:embed stub.geojson
var stub []byte

func main() {

	year := flag.Int("year", 0, "The year for the new map.")
	name := flag.String("name", "", "An optional name to assign. If empty name will be \"SFO ({YEAR})\"")

	inception := flag.String("inception", "", "An optional inception date for the new map.")

	writer_uri := flag.String("writer-uri", "repo:///usr/local/data/sfomuseum-data-maps", "A valid whosonfirst/go-writer URI where the new map will be written.")

	parent_reader_uri := flag.String("parent-reader-uri", "repo:///usr/local/data/sfomuseum-data-whosonfirst", "A valid whosonfirst/go-reader URI to use for reading data for -parent-id.")
	parent_id := flag.Int64("parentid", 102527513, "SFO")

	var depicts multi.MultiInt64
	flag.Var(&depicts, "depicts", "Zero or more WOF (sfomuseum-data) IDs that this map depicts")

	flag.Parse()

	if *year == 0 {
		log.Fatalf("Invalid year")
	}

	if *name == "" {
		*name = fmt.Sprintf("SFO (%d)", *year)
	}

	ctx := context.Background()

	wr, err := writer.NewWriter(ctx, *writer_uri)

	if err != nil {
		log.Fatalf("Failed to create new writer, %v", err)
	}

	parent_r, err := reader.NewReader(ctx, *parent_reader_uri)

	if err != nil {
		log.Fatalf("Failed to create parent reader, %v", err)
	}

	p_body, err := wof_reader.LoadBytes(ctx, parent_r, *parent_id)

	if err != nil {
		log.Fatalf("Failed to load parent record, %v", err)
	}

	p_hier := properties.Hierarchies(p_body)

	p_geom, err := geometry.Geometry(p_body)

	if err != nil {
		log.Fatalf("Failed to derive parent geometry, %v", err)
	}

	orb_geom := p_geom.Geometry()
	p_bounds := orb_geom.Bound()

	poly := p_bounds.ToPolygon()

	geom := geojson.NewGeometry(poly)

	updates := map[string]interface{}{
		"properties.sfomuseum:uri": *year,
		"properties.wof:name":      *name,
		"properties.wof:hierarchy": p_hier,
		"geometry":                 geom,
	}

	if *inception != "" {
		updates["properties.edtf:inception"] = *inception
		updates["properties.edtf:cessation"] = ".."
	}

	if len(depicts) > 0 {
		updates["properties.wof:depicts"] = depicts
	}

	body, err := export.AssignProperties(ctx, stub, updates)

	if err != nil {
		log.Fatalf("Failed to assign properties, %v", err)
	}

	new_id, err := sfom_writer.WriteBytes(ctx, wr, body)

	if err != nil {
		log.Fatalf("Failed to write bytes, %v", err)
	}

	fmt.Println(new_id)
}
