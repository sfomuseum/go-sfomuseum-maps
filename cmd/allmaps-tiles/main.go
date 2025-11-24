// `allmaps-tiles` automates (most of) the process of generating ZXY tiles from an Allmaps annotations URL in an SFO Museum context.
// It is NOT a general purpose tool. As of this writing it also writes all its files to the current working directory and depends on
// a number of other tools to already be installed: the allmaps command line tool, gdal, curl. Final output files are not moved, renamed
// or automatically cleaned up. They should be eventually but today they are not.
//
// Note: As of March 2025, Helmert transformations which are available in the Allmaps Editor UI are NOT available in GDAL warp/translate
// operations. Womp womp...
package main

/*

$> go run cmd/allmaps-tiles/main.go -map-url https://annotations.allmaps.org/maps/d774c394c460ccfe
2025/03/11 10:51:52 INFO Identifiers annotation=d774c394c460ccfe image=c157f0e8c25aa123 sfomuseum=1762891451
2025/03/11 10:51:52 INFO Source image url=https://static.sfomuseum.org/media/176/289/145/1/1762891451_ghb7Nar9CJ8ylb8tJ4GsePOolOqn0lOG_o.jpg
2025/03/11 10:51:52 INFO Allmaps image path=/usr/local/sfomuseum/go-sfomuseum-maps/c157f0e8c25aa123.jpg
2025/03/11 10:51:52 INFO Warped image path=/usr/local/sfomuseum/go-sfomuseum-maps/c157f0e8c25aa123_d774c394c460ccfe-warped.tif
2025/03/11 10:51:52 INFO Allmaps script path=/usr/local/sfomuseum/go-sfomuseum-maps/c157f0e8c25aa123.sh

$> ll c157f0e8c25aa123_d774c394c460ccfe*
-rw-r--r--  1 asc  staff  556935 Mar 11 10:51 c157f0e8c25aa123_d774c394c460ccfe-warped.tif
-rw-r--r--  1 asc  staff     715 Mar 11 10:51 c157f0e8c25aa123_d774c394c460ccfe.geojson
-rw-r--r--  1 asc  staff    2137 Mar 11 10:51 c157f0e8c25aa123_d774c394c460ccfe.vrt

c157f0e8c25aa123_d774c394c460ccfe-warped:
total 96
drwxr-xr-x  3 asc  staff     96 Mar 11 10:44 12
drwxr-xr-x  4 asc  staff    128 Mar 11 10:44 13
drwxr-xr-x  5 asc  staff    160 Mar 11 10:44 14
drwxr-xr-x  7 asc  staff    224 Mar 11 10:44 15
-rw-r--r--  1 asc  staff  27890 Mar 11 10:51 googlemaps.html
-rw-r--r--@ 1 asc  staff   4214 Mar 11 10:51 leaflet.html
-rw-r--r--  1 asc  staff   4763 Mar 11 10:51 openlayers.html
-rw-r--r--  1 asc  staff    879 Mar 11 10:51 tilemapresource.xml

*/

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jtacoma/uritemplates"
	"github.com/sfomuseum/go-sfomuseum-maps/allmaps"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-reader/v2"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader/v2"
)

func main() {

	var map_url string
	var media_reader_uri string

	flag.StringVar(&map_url, "map-url", "", "The Allmaps annotations URL of the map you want to tile.")
	flag.StringVar(&media_reader_uri, "media-reader-uri", "repo:///usr/local/data/sfomuseum-data-media-collection", "A registered whosonfirst/go-reader.Reader URI to use for retrieving SFO Museum media (image) data.")

	flag.Parse()

	ctx := context.Background()

	media_r, err := reader.NewReader(ctx, media_reader_uri)

	if err != nil {
		log.Fatalf("Failed to create media reader, %v", err)
	}

	annotation_id, err := allmaps.DeriveAnnotationId(map_url)

	if err != nil {
		log.Fatalf("Failed to derive Allmaps annotation ID, %v", err)
	}

	image_id, err := allmaps.DeriveImageId(map_url)

	if err != nil {
		log.Fatalf("Failed to derive Allmaps image ID, %v", err)
	}

	sfom_id, err := allmaps.DeriveSFOMuseumImageId(map_url)

	if err != nil {
		log.Fatalf("Failed to derive SFO Museum image ID, %v", err)
	}

	slog.Info("Identifiers", "annotation", annotation_id, "image", image_id, "sfomuseum", sfom_id)

	cwd, err := os.Getwd()

	if err != nil {
		log.Fatalf("Failed to derive current working directory, %v", err)
	}

	// START OF derive (SFO Museum) image URL to rectify
	// START OF put me in a function or something...

	media_body, err := wof_reader.LoadBytes(ctx, media_r, sfom_id)

	if err != nil {
		log.Fatalf("Failed to read body for %d, %v", sfom_id, err)
	}

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

	slog.Info("Source image", "url", image_url)

	// END OF put me in a function or something...
	// END OF derive (SFO Museum) image URL to rectify

	// START OF copy source image to {ALLMAPS_IMAGE_ID}.jpg

	rsp, err := http.Get(image_url)

	if err != nil {
		log.Fatalf("Failed to fetch %s, %v", image_url, err)
	}

	defer rsp.Body.Close()

	allmaps_im := fmt.Sprintf("%s.jpg", image_id)
	allmaps_sh := fmt.Sprintf("%s.sh", image_id)
	warped_im := fmt.Sprintf("%s_%s-warped.tif", image_id, annotation_id)

	allmaps_im = filepath.Join(cwd, allmaps_im)
	allmaps_sh = filepath.Join(cwd, allmaps_sh)
	warped_im = filepath.Join(cwd, warped_im)

	slog.Info("Allmaps image", "path", allmaps_im)
	slog.Info("Warped image", "path", warped_im)
	slog.Info("Allmaps script", "path", allmaps_sh)

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

	defer os.Remove(allmaps_im)

	// END OF copy source image to {ALLMAPS_IMAGE_ID}.jpg

	// START OF do all the gdal things
	// In principal we could generate allmaps.sh ourselves removing
	// the need for node and allmaps-cli but then we have to track
	// and reflect all the changes to the latter so...

	// It is probably worth creating a Golang struct reflecting the Allmaps
	// annotations data format and using that (and the GCPs it defines) to
	// generate the gdal_translate command. But not today...

	sh, err := os.OpenFile(allmaps_sh, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		log.Fatalf("Failed to open %s for writing, %v", allmaps_sh, err)
	}

	l1 := fmt.Sprintf("curl -s %s | allmaps script geotiff | bash\n", map_url)
	l2 := fmt.Sprintf("gdal2tiles.py %s\n", warped_im)

	sh.Write([]byte(l1))
	sh.Write([]byte(l2))

	err = sh.Close()

	if err != nil {
		log.Fatalf("Failed to close %s after writing, %v", allmaps_sh, err)
	}

	defer os.Remove(allmaps_sh)

	cmd := exec.Command("sh", allmaps_sh)
	err = cmd.Run()

	if err != nil {
		log.Fatalf("Failed to run %s command, %v", allmaps_sh, err)
	}

	// END OF do all the gdal things
}
