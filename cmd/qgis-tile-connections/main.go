package main

import (
	_ "github.com/whosonfirst/go-whosonfirst-iterate-git/v3"
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
	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
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

	iterator_uri := flag.String("iterator-uri", "git://", "A valid whosonfirst/go-whosonfirst-iterate-git/v2/iterator URI.")
	iterator_source := flag.String("iterator-source", "https://github.com/sfomuseum-data/sfomuseum-data-maps.git", "A valid whosonfirst/go-whosonfirst-iterate-git/v2/iterator source.")

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

	iter, err := iterate.NewIterator(ctx, *iterator_uri)

	if err != nil {
		log.Fatalf("Failed to create new iterator, %v", err)
	}

	for rec, err := range iter.Iterate(ctx, *iterator_source) {

		if err != nil {
			log.Fatalf("Iterator yielded an error, %v", err)
		}

		defer rec.Body.Close()

		if filepath.Ext(rec.Path) != ".geojson" {
			continue
		}

		body, err := io.ReadAll(rec.Body)

		if err != nil {
			log.Fatal(err)
		}

		label, err := maps.DeriveLabel(body)

		if err != nil {
			log.Fatalf("Failed to derive year label for %s, %v", rec.Path, err)
		}

		if len(exclude) > 0 {

			for _, e := range exclude {
				if label == e {
					continue
				}
			}
		}

		year_label, err := maps.DeriveYearLabel(body)

		if err != nil {
			log.Fatalf("Failed to derive year label for %s, %v", rec.Path, err)
		}

		min_zoom, max_zoom, err := maps.DeriveZoomLevels(body)

		if err != nil {
			log.Fatalf("Failed to derive zoom levels for %s, %v", rec.Path, err)
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

		_, exists := tile_dict[t.label]

		if exists {
			log.Fatalf("Duplicate label %s", t.label)
		}

		tile_dict[t.label] = t
	}

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
