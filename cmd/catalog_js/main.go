package main

import (
	_ "github.com/whosonfirst/go-whosonfirst-iterate-git/v2"
)

import (
	"context"
	"encoding/json"
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
	"github.com/sfomuseum/go-sfomuseum-maps/templates/javascript"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
)

type MapDict map[string]Map
type MapCatalog []Map

type Map struct {
	Label string `json:"label"`
	Year  int    `json:"year"` // Deprecated
	// Date       string `json:"date"`
	MinZoom    int    `json:"min_zoom"`
	MaxZoom    int    `json:"max_zoom"`
	Source     string `json:"source,omitempty"`
	Identifier string `json:"identifier,omitempty"`
	URL        string `json:"url"`
}

type TemplateVars struct {
	Catalog      MapCatalog
	LastModified string
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

	t := template.New("sfomuseum_maps").Funcs(template.FuncMap{
		"ToJSON": func(raw interface{}) string {

			enc, err := json.Marshal(raw)

			if err != nil {
				log.Println(err)
				return ""
			}

			return string(enc)
		},
	})

	t, err := t.ParseFS(javascript.FS, "*.js")

	if err != nil {
		log.Fatal(err)
	}

	t = t.Lookup("catalog_js")

	if t == nil {
		log.Fatal("Missing catalog")
	}

	map_dict := make(map[string]Map)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	map_ch := make(chan Map)
	done_ch := make(chan bool)

	cb := func(ctx context.Context, path string, fh io.ReadSeeker, args ...interface{}) error {

		if filepath.Ext(path) != ".geojson" {
			return nil
		}

		body, err := io.ReadAll(fh)

		if err != nil {
			return fmt.Errorf("Failed to read data for %s, %w", path, err)
		}

		src_rsp := gjson.GetBytes(body, "properties.src:geom")

		if !src_rsp.Exists() {
			return fmt.Errorf("%s is missing src:geom", path)
		}

		src := src_rsp.String()

		path_id := fmt.Sprintf("properties.%s:id", src)

		id_rsp := gjson.GetBytes(body, path_id)
		src_id := ""

		if id_rsp.Exists() {
			src_id = id_rsp.String()
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

		// START OF deprecate this...

		incept_rsp := gjson.GetBytes(body, "properties.date:inception_upper")

		if !incept_rsp.Exists() {
			return fmt.Errorf("%s is missing date:inception_upper", path)
		}

		incept_str := incept_rsp.String()
		incept_t, err := time.Parse("2006-01-02", incept_str)

		if err != nil {
			return err
		}

		year := incept_t.Year()

		// END OF deprecate this...

		min_zoom, max_zoom, err := maps.DeriveZoomLevels(body)

		if err != nil {
			return fmt.Errorf("Failed to derive zoom levels for %s, %w", path, err)
		}

		url := fmt.Sprintf("https://static.sfomuseum.org/aerial/%s/{z}/{x}/{-y}.png", label)

		m := Map{
			Label: label,
			Year:  year,
			// Date:       year_label,
			MinZoom:    min_zoom,
			MaxZoom:    max_zoom,
			Source:     src,
			Identifier: src_id,
			URL:        url,
		}

		map_ch <- m
		return nil
	}

	go func() {

		for {
			select {
			case <-ctx.Done():
				return
			case <-done_ch:
				return
			case m := <-map_ch:

				_, exists := map_dict[m.Label]

				if exists {
					log.Fatalf("Duplicate Label")
				}

				map_dict[m.Label] = m
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

	for label, _ := range map_dict {
		labels = append(labels, label)
	}

	sort.Strings(labels)

	map_catalog := make([]Map, 0)

	for _, label := range labels {
		map_catalog = append(map_catalog, map_dict[label])
	}

	now := time.Now()

	vars := TemplateVars{
		Catalog:      map_catalog,
		LastModified: now.Format(time.RFC3339),
		Args:         strings.Join(os.Args, " "),
	}

	out := os.Stdout
	err = t.Execute(out, vars)

	if err != nil {
		log.Fatalf("Failed to execute template, %v", err)
	}

}
