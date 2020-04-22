package main

import (
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-index"
	_ "github.com/whosonfirst/go-whosonfirst-index/fs"	
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {

	repo := flag.String("repo", "/usr/local/data/sfomuseum-data-maps", "...")

	flag.Parse()

	wr := os.Stdout	
	csv_wr := csv.NewWriter(wr)

	mu := new(sync.RWMutex)
	
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

		bbox_rsp := gjson.GetBytes(body, "bbox")

		if !bbox_rsp.Exists() {
			return errors.New("Missing bbox")
		}

		incept_rsp := gjson.GetBytes(body, "properties.date:inception_upper")

		if !incept_rsp.Exists() {
			return errors.New("Missing date:inception_upper")
		}

		min_zoom := int(min_rsp.Int())
		max_zoom := int(max_rsp.Int())

		bbox_array := bbox_rsp.Array()

		if len(bbox_array) != 4 {
			return errors.New("Weird bbox")
		}

		bbox := []string{
			bbox_array[1].String(),
			bbox_array[0].String(),
			bbox_array[3].String(),
			bbox_array[2].String(),
		}

		incept_str := incept_rsp.String()
		incept_t, err := time.Parse("2006-01-02", incept_str)

		if err != nil {
			return err
		}

		year := incept_t.Year()
		str_year := strconv.Itoa(year)

		if str_year == "" {
			return nil
		}
		
		str_bbox := strings.Join(bbox, ",")

		zooms := make([]string, 0)

		for i := min_zoom; i <= max_zoom; i++ {
			zooms = append(zooms, strconv.Itoa(i))
		}

		str_zooms := strings.Join(zooms, ",")
		
		out := []string{
			str_year,
			// strconv.Itoa(min_zoom),
			// strconv.Itoa(max_zoom),
			str_zooms,
			str_bbox,
		}

		mu.Lock()
		defer mu.Unlock()
		
		csv_wr.Write(out)
		
		return nil
	}

	idx, err := index.NewIndexer("repo", cb)

	if err != nil {
		log.Fatal(err)
	}

	err = idx.IndexPath(*repo)

	if err != nil {
		log.Fatal(err)
	}

	csv_wr.Flush()	
}
