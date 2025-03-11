package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/jtacoma/uritemplates"
	"github.com/sfomuseum/go-sfomuseum-maps/allmaps"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-reader"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader"
)

func main() {

	var map_url string

	var media_reader_uri string

	flag.StringVar(&map_url, "map-url", "", "...")
	flag.StringVar(&media_reader_uri, "media-reader-uri", "repo:///usr/local/data/sfomuseum-data-media-collection", "...")

	flag.Parse()

	ctx := context.Background()

	media_r, err := reader.NewReader(ctx, media_reader_uri)

	if err != nil {
		log.Fatalf("Failed to create media reader, %v", err)
	}

	allmaps_id, err := allmaps.DeriveImageId(map_url)

	if err != nil {
		log.Fatalf("Failed to derive Allmaps image ID, %v", err)
	}

	sfom_id, err := allmaps.DeriveSFOMuseumImageId(map_url)

	if err != nil {
		log.Fatalf("Failed to derive SFO Museum image ID, %v", err)
	}

	log.Println(map_url, allmaps_id, sfom_id)

	media_body, err := wof_reader.LoadBytes(ctx, media_r, sfom_id)

	if err != nil {
		log.Fatalf("Failed to read body for %d, %v", sfom_id, err)
	}

	// START OF put me in a function or something...

	sz_label := "o"
	sz_path := fmt.Sprintf("properties.media:properties.sizes.%s", sz_label)

	sz_rsp := gjson.GetBytes(media_body, sz_path)

	if !sz_rsp.Exists() {
		log.Fatalf("Missing image size property")
	}

	t_rsp := gjson.GetBytes(media_body, "properties.media:uri_template")

	if !t_rsp.Exists() {
		log.Fatalf("Missing URI template property")
	}

	secret_rsp := sz_rsp.Get("secret")
	ext_rsp := sz_rsp.Get("extension")

	t, err := uritemplates.Parse(t_rsp.String())

	if err != nil {
		log.Fatalf("Failed to parse URI template, %v", err)
	}

	vars := map[string]interface{}{
		"label":     sz_label,
		"secret":    secret_rsp.String(),
		"extension": ext_rsp.String(),
	}

	image_url, err := t.Expand(vars)

	if err != nil {
		log.Fatalf("Failed to expand URI template, %v", err)
	}

	log.Println(image_url)

	// END OF put me in a function or something...
}
