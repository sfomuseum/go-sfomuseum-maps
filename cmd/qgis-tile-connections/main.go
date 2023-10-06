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
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/sfomuseum/go-flags/multi"
	"github.com/sfomuseum/go-sfomuseum-maps/templates/xml"
	"github.com/tidwall/gjson"
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

	// mode := flag.String("mode", "repo://", "...")
	// uri := flag.String("uri", "/usr/local/data/sfomuseum-data-maps", "...")

	mode := flag.String("mode", "git://", "...")
	uri := flag.String("uri", "https://github.com/sfomuseum-data/sfomuseum-data-maps.git", "...")

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

		min_rsp := gjson.GetBytes(body, "properties.mz:min_zoom")

		if !min_rsp.Exists() {
			return fmt.Errorf("%s is missing mz:min_zoom", path)
		}

		max_rsp := gjson.GetBytes(body, "properties.mz:max_zoom")

		if !max_rsp.Exists() {
			return fmt.Errorf("%s is missing mz:max_zoom", path)
		}

		uri_rsp := gjson.GetBytes(body, "properties.sfomuseum:uri")

		if !uri_rsp.Exists() {
			return fmt.Errorf("Missing sfomuseum:uri")
		}

		label := uri_rsp.String()

		if len(exclude) > 0 {

			for _, e := range exclude {
				if label == e {
					return nil
				}
			}
		}

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

		min_zoom := int(min_rsp.Int())
		max_zoom := int(max_rsp.Int())

		url := fmt.Sprintf("https://static.sfomuseum.org/aerial/%s/{z}/{x}/{-y}.png", label)

		t := &ZXYTiles{
			Name:           fmt.Sprintf("SFO %s (SFO Museum)", label),
			label:          strconv.Itoa(year),
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
