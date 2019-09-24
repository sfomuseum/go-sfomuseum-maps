package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-index"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
	"text/template"
)

type Map struct {
	Year int	`json:"year"`
	MinZoom int	`json:"min_zoom"`
	MaxZoom int	`json:"max_zoom"`
	Source int	`json:"source,omitempty"`
	Identifier string	`json:"identifier,omitempty"`
}
	
type TemplateVars struct {
	Maps string
}

func main() {

	repo := flag.String("repo", "/usr/local/data/sfomuseum-data-maps", "...")
	path_templates := flag.String("templates", "", "...")
	
	flag.Parse()

	t := template.New("sfomuseum_maps").Funcs(template.FuncMap{
	})
	
	t, err := t.ParseGlob(*path_templates)

	if err != nil {
		log.Fatal(err)
	}
	
	t = t.Lookup("catalog_js")

	if t == nil {
		log.Fatal("Missing catalog")
	}
	
	maps := make([]Map, 0)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	map_ch := make(chan Map)
	done_ch := make(chan bool)
	
	cb := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

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

		bbox_rsp := gjson.GetBytes(body, "bbox")

		if !bbox_rsp.Exists() {
			return errors.New("Missing bbox")
		}

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
			Year: year,
			MinZoom: min_zoom,
			MaxZoom: max_zoom,
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
				maps = append(maps, m)
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

	enc_maps, err := json.Marshal(maps)

	if err != nil {
		log.Fatal(err)
	}

	str_maps := string(enc_maps)
	
	vars := TemplateVars{
		Maps: str_maps,
	}

	out := os.Stdout
	err = t.Execute(out, vars)

	if err != nil {
		log.Fatal(err)
	}

}


