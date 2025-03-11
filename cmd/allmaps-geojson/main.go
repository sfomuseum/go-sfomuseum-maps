// `allmaps-geojson` automates (most of) the process of generating a sfomuseum-data-maps GeoJSON file derive from an Allmaps annotations GeoJSON URL.
// It is NOT a general purpose tool. 
package main

/*

> go run cmd/allmaps-geojson/main.go -writer-uri stdout:// -map-url https://annotations.allmaps.org/images/c157f0e8c25aa123.geojson -min-zoom 12 -max-zoom 15 -year 1936 -name 'Mills Field (1936)'

*/

import (
	"context"
	"flag"
	"io"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/paulmach/orb/geojson"
	"github.com/whosonfirst/go-writer/v3"
	wof_id "github.com/whosonfirst/go-whosonfirst-id"	
	wof_writer "github.com/whosonfirst/go-whosonfirst-writer/v3"	
)

func main() {

	var map_url string
	var writer_uri string

	var name string
	var parent_id int64
	var year int

	var min_zoom int
	var max_zoom int	
	
	flag.StringVar(&map_url, "map-url", "", "...")
	flag.StringVar(&writer_uri, "writer-uri", "repo:///usr/local/data/sfomuseum-data-maps", "...")

	flag.StringVar(&name, "name", "", "...")
	flag.Int64Var(&parent_id, "parent-id", -1, "...")
	flag.IntVar(&year, "year", 0, "...")
	flag.IntVar(&min_zoom, "min-zoom", 0, "...")
	flag.IntVar(&max_zoom, "max-zoom", 0, "...")	
	
	flag.Parse()

	ctx := context.Background()

	wr, err := writer.NewWriter(ctx, writer_uri)

	if err != nil {
		log.Fatalf("Failed to create writer, %v", err)
	}

	// START OF put me in a function... ?
	
	rsp, err := http.Get(map_url)

	if err != nil {
		log.Fatalf("Failed to retrieve %s, %v", err)
	}

	defer rsp.Body.Close()

	body, err := io.ReadAll(rsp.Body)

	if err != nil {
		log.Fatalf("Failed to read %s, %v", map_url, err)
	}

	fc, err := geojson.UnmarshalFeatureCollection(body)

	if err != nil {
		log.Fatalf("Failed to unmarshal %s, %v", map_url, err)
	}

	if len(fc.Features) != 1 {
		log.Fatalf("Invalid feature count")
	}

	f := fc.Features[0]

	// END OF put me in a function... ?

	new_id, err := wof_id.NewID()

	if err != nil {
		log.Fatalf("Failed to derive new ID, %v", err)
	}
	
	new_props := map[string]any{
		"wof:id": new_id,
		"wof:name": name,
		"wof:repo": "sfomuseum-data-maps",
		"wof:placetype": "custom",
		"wof:placetype_alt": []string{
			"map",
		},
		"wof:country": "US",
		"wof:parent_id": parent_id,
		"sfomuseum:placetype": "map",
		"sfomuseum:uri": year,
		"mz:is_current": -1,
		"mz:max_zoom": max_zoom,
		"mz:min_zoom": min_zoom,
		"edtf:inception": year,
		"edtf:cessation": year,
		"src:geom": "sfomuseum",
	}

	// If parent_id != -1 then: get hierarchy...
	
	for k, v := range f.Properties {

		if k == "_allmaps" {
			k = "meta"
		}
		
		new_k := fmt.Sprintf("allmaps:%s", k)
		new_props[new_k] = v
	}

	f.Properties = new_props

	enc_f, err := f.MarshalJSON()

	if err != nil {
		log.Fatalf("Failed to marshal JSON, %v", err)
	}

	_, err = wof_writer.WriteBytes(ctx, wr, enc_f)

	if err != nil {
		log.Fatalf("Failed to write data, %v", err)
	}

	slog.Info("Created new record", "id", new_id)
}

