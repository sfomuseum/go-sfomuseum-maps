package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-index"
	_ "github.com/whosonfirst/go-whosonfirst-index/fs"	
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
	"text/template"
)

type MapCatalog map[string]Map

type Map struct {
	URI string `json:"uri"`
	Year int	`json:"year"`
	MinZoom int	`json:"min_zoom"`
	MaxZoom int	`json:"max_zoom"`
	Source string	`json:"source,omitempty"`
	Identifier string	`json:"identifier,omitempty"`
}
	
type TemplateVars struct {
	Catalog MapCatalog
	LastModified string
}

func main() {

	repo := flag.String("repo", "/usr/local/data/sfomuseum-data-maps", "...")
	path_templates := flag.String("templates", "", "...")
	
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
	
	t, err := t.ParseGlob(*path_templates)

	if err != nil {
		log.Fatal(err)
	}
	
	t = t.Lookup("catalog_js")

	if t == nil {
		log.Fatal("Missing catalog")
	}
	
	map_catalog := make(map[string]Map)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	map_ch := make(chan Map)
	done_ch := make(chan bool)
	
	cb := func(ctx context.Context, fh io.Reader, args ...interface{}) error {

		body, err := ioutil.ReadAll(fh)

		if err != nil {
			return err
		}

		min_rsp := gjson.GetBytes(body, "properties.mz:min_zoom")

		if !min_rsp.Exists() {
			return errors.New("Missing mz:min_zoom")
		}

		max_rsp := gjson.GetBytes(body, "properties.mz:max_zoom")

		if !max_rsp.Exists() {
			return errors.New("Missing mz:max_zoom")
		}

		src_rsp := gjson.GetBytes(body, "properties.src:geom")

		if !src_rsp.Exists() {
			return errors.New("Missing src:geom")
		}

		src := src_rsp.String()

		path_id := fmt.Sprintf("properties.%s:id", src)

		id_rsp := gjson.GetBytes(body, path_id)
		src_id := ""
			
		if id_rsp.Exists() {
			src_id = id_rsp.String()
		}

		
		uri_rsp := gjson.GetBytes(body, "properties.sfomuseum:uri")

		if !uri_rsp.Exists() {
			return errors.New("Missing sfomuseum:uri")
		}

		uri := uri_rsp.String()
		
		incept_rsp := gjson.GetBytes(body, "properties.date:inception_upper")

		if !incept_rsp.Exists() {
			return errors.New("Missing date:inception_upper")
		}

		incept_str := incept_rsp.String()
		incept_t, err := time.Parse("2006-01-02", incept_str)

		if err != nil {
			return err
		}

		year := incept_t.Year()

		min_zoom := int(min_rsp.Int())
		max_zoom := int(max_rsp.Int())

		m := Map{
			URI: uri,
			Year: year,
			MinZoom: min_zoom,
			MaxZoom: max_zoom,
			Source: src,
			Identifier: src_id,			
		}

		map_ch <- m
		return nil
	}

	go func() {

		for {
			select {
			case <- ctx.Done():
				return				
			case <- done_ch:
				return
			case m := <- map_ch:

				_, exists := map_catalog[m.URI]

				if exists {
					log.Fatalf("Duplicate URI")
				}
				
				map_catalog[m.URI] = m
			}
		}

	}()
	
	idx, err := index.NewIndexer("repo", cb)

	if err != nil {
		log.Fatal(err)
	}

	err = idx.IndexPath(*repo)

	if err != nil {
		log.Fatal(err)
	}

	done_ch <- true

	now := time.Now()
	
	vars := TemplateVars{
		Catalog: map_catalog,
		LastModified: now.Format(time.RFC3339),
	}

	out := os.Stdout
	err = t.Execute(out, vars)

	if err != nil {
		log.Fatal(err)
	}

}


