package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/paulmach/orb/geojson"
	wof_writer "github.com/whosonfirst/go-whosonfirst-writer/v3"
	"github.com/whosonfirst/go-writer/v3"
)

func main() {

	var map_url string
	var writer_uri string

	flag.StringVar(&map_url, "map-url", "", "...")
	flag.StringVar(&writer_uri, "writer-uri", "repo:///usr/local/data/sfomuseum-data-maps", "...")

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

	new_props := map[string]any{
		"wof:name":      "",
		"wof:repo":      "sfomuseum-data-maps",
		"wof:placetype": "custom",
		"wof:placetype_alt": []string{
			"map",
		},
		"wof:country":         "US",
		"wof:parent_id":       -1,
		"sfomuseum:placetype": "map",
	}

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

}
