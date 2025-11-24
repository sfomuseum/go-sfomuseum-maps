package main

import (
	_ "github.com/whosonfirst/go-whosonfirst-iterate-git/v3"
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
	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
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

	iterator_uri := flag.String("iterator-uri", "git://", "A valid whosonfirst/go-whosonfirst-iterate-git/v3.Iterator URI.")
	iterator_source := flag.String("iterator-source", "https://github.com/sfomuseum-data/sfomuseum-data-maps.git", "A valid whosonfirst/go-whosonfirst-iterate-git/v3.Iterator source.")

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

	iter, err := iterate.NewIterator(ctx, *iterator_uri)

	if err != nil {
		log.Fatalf("Failed to create new iterator, %v", err)
	}

	for rec, err := range iter.Iterate(ctx, *iterator_source) {

		if err != nil {
			log.Fatalf("Iterator reported an error, %v", err)
		}

		defer rec.Body.Close()

		if filepath.Ext(rec.Path) != ".geojson" {
			continue
		}

		body, err := io.ReadAll(rec.Body)

		if err != nil {
			log.Fatalf("Failed to read data for %s, %v", rec.Path, err)
		}

		src_rsp := gjson.GetBytes(body, "properties.src:geom")

		if !src_rsp.Exists() {
			log.Fatalf("%s is missing src:geom", rec.Path)
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
			log.Fatalf("Failed to derive year label for %s, %v", rec.Path, err)
		}

		if len(exclude) > 0 {

			for _, e := range exclude {
				if label == e {
					continue
				}
			}
		}

		// START OF deprecate this...

		incept_rsp := gjson.GetBytes(body, "properties.date:inception_upper")

		if !incept_rsp.Exists() {
			log.Fatalf("%s is missing date:inception_upper", rec.Path)
		}

		incept_str := incept_rsp.String()
		incept_t, err := time.Parse("2006-01-02", incept_str)

		if err != nil {
			log.Fatal(err)
		}

		year := incept_t.Year()

		// END OF deprecate this...

		min_zoom, max_zoom, err := maps.DeriveZoomLevels(body)

		if err != nil {
			log.Fatalf("Failed to derive zoom levels for %s, %v", rec.Path, err)
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

		_, exists := map_dict[m.Label]

		if exists {
			log.Fatalf("Duplicate label, %s", m.Label)
		}

		map_dict[m.Label] = m
	}

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
