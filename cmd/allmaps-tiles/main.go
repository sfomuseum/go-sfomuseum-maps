package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

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

	rsp, err := http.Get(image_url)

	if err != nil {
		log.Fatalf("Failed to fetch %s, %v", image_url, err)
	}

	defer rsp.Body.Close()

	allmaps_im := fmt.Sprintf("%s.jpg", allmaps_id)

	wr, err := os.OpenFile(allmaps_im, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		log.Fatalf("Failed to open %s for writing, %v", allmaps_im, err)
	}

	_, err = io.Copy(wr, rsp.Body)

	if err != nil {
		log.Fatalf("Failed to copy %s to %s, %v", image_url, allmaps_im, err)
	}

	err = wr.Close()

	if err != nil {
		log.Fatalf("Failed to close %s after writing, %v", allmaps_im, err)
	}

	/*

		$> curl -s https://annotations.allmaps.org/maps/a0c0c652e49f4596 | allmaps script geotiff | bash
		$> gdal2tiles.py ced8faec8c108002_a0c0c652e49f4596-warped.tif

	*/
}
