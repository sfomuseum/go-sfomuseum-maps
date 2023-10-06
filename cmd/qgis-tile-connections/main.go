package main

import (
	_ "github.com/whosonfirst/go-whosonfirst-iterate-git/v2"
)

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/sfomuseum/go-flags/multi"
	"github.com/sfomuseum/go-sfomuseum-maps"
	"github.com/sfomuseum/go-sfomuseum-maps/templates/xml"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
)

type ZXYTiles struct {
	Username       string
	Password       string
	ZMin           int
	ZMax           int
	TilePixelRatio int
	URL            string
	Name           string
	AuthConfig     string
	Referer        string
	label          string
}

type TemplateVars struct {
	TileConnections []*ZXYTiles
	LastModified    string
	// as in the command-line args so we can understand how the file was create
	// this was largely to account for the desire to exclude 1937 from the T2 installation
	Args string
}

func main() {

	mode := flag.String("mode", "git://", "A valid whosonfirst/go-whosonfirst-iterate-git/v2/iterator URI.")
	uri := flag.String("uri", "https://github.com/sfomuseum-data/sfomuseum-data-maps.git", "A valid whosonfirst/go-whosonfirst-iterate-git/v2/iterator source.")

	var exclude multi.MultiString
	flag.Var(&exclude, "exclude", "Zero or more maps to exclude (based on their sfomuseum:uri value)")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t := template.New("sfomuseum_maps") // .Funcs(template.FuncMap{})

	t, err := t.ParseFS(xml.FS, "*.xml")

	if err != nil {
		log.Fatal(err)
	}

	t = t.Lookup("qgis_tile_connections")

	if t == nil {
		log.Fatal("Missing template")
	}

	tile_dict := make(map[string]*ZXYTiles)

	tile_ch := make(chan *ZXYTiles)
	done_ch := make(chan bool)

	cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

		if filepath.Ext(path) != ".geojson" {
			return nil
		}

		body, err := io.ReadAll(r)

		if err != nil {
			return err
		}

		label, err := maps.DeriveLabel(body)

		if err != nil {
			return fmt.Errorf("Failed to derive year label for %s, %w", path, err)
		}

		if len(exclude) > 0 {

			for _, e := range exclude {
				if label == e {
					return nil
				}
			}
		}

		year_label, err := maps.DeriveYearLabel(body)

		if err != nil {
			return fmt.Errorf("Failed to derive year label for %s, %w", path, err)
		}

		min_zoom, max_zoom, err := maps.DeriveZoomLevels(body)

		if err != nil {
			return fmt.Errorf("Failed to derive zoom levels for %s, %w", path, err)
		}

		url := fmt.Sprintf("https://static.sfomuseum.org/aerial/%s/{z}/{x}/{-y}.png", label)

		name := fmt.Sprintf("SFO %s (SFO Museum)", label)

		t := &ZXYTiles{
			Name:           name,
			label:          year_label,
			ZMin:           min_zoom,
			ZMax:           max_zoom,
			URL:            url,
			TilePixelRatio: 1,
		}

		tile_ch <- t
		return nil
	}

	go func() {

		for {
			select {
			case <-ctx.Done():
				return
			case <-done_ch:
				return
			case t := <-tile_ch:

				_, exists := tile_dict[t.label]

				if exists {
					log.Fatalf("Duplicate label")
				}

				tile_dict[t.label] = t
			}
		}

	}()

	iter, err := iterator.NewIterator(ctx, *mode, cb)

	if err != nil {
		log.Fatalf("Failed to create new iterator, %v", err)
	}

	err = iter.IterateURIs(ctx, *uri)

	if err != nil {
		log.Fatalf("Failed to iterate '%s', %v", *uri, err)
	}

	done_ch <- true

	labels := make([]string, 0)

	for label, _ := range tile_dict {
		labels = append(labels, label)
	}

	sort.Strings(labels)

	tile_connections := make([]*ZXYTiles, 0)

	for _, label := range labels {
		tile_connections = append(tile_connections, tile_dict[label])
	}

	now := time.Now()

	vars := TemplateVars{
		TileConnections: tile_connections,
		LastModified:    now.Format(time.RFC3339),
		Args:            strings.Join(os.Args, " "),
	}

	out := os.Stdout
	err = t.Execute(out, vars)

	if err != nil {
		log.Fatalf("Failed to execute template, %v", err)
	}

}
