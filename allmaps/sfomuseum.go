package allmaps

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

func DeriveSFOMuseumImageId(allmaps_url string) (int64, error) {

	u, err := url.Parse(allmaps_url)

	if err != nil {
		return -1, err
	}

	if u.Host != "annotations.allmaps.org" {
		return -1, fmt.Errorf("Invalid host")
	}

	var tiles_path string

	// https://annotations.allmaps.org/images/c157f0e8c25aa123.geojson

	if strings.HasPrefix(u.Path, "/images/") {
		tiles_path = "features.0.properties.resource.uri"
	}

	// https://annotations.allmaps.org/maps/a0c0c652e49f4596

	if strings.HasPrefix(u.Path, "/maps/") {
		tiles_path = "target.source.id"
	}

	if tiles_path == "" {
		return -1, fmt.Errorf("Failed to derive tiles path")
	}

	rsp, err := http.Get(allmaps_url)

	if err != nil {
		return -1, err
	}

	defer rsp.Body.Close()

	body, err := io.ReadAll(rsp.Body)

	if err != nil {
		return -1, fmt.Errorf("Failed to read body, %w", err)
	}

	tiles_rsp := gjson.GetBytes(body, tiles_path)

	if !tiles_rsp.Exists() {
		return -1, fmt.Errorf("Body missing tile data")
	}

	tiles_uri := tiles_rsp.String()

	re_tiles, err := regexp.Compile(`https:\/\/static\.sfomuseum\.org\/media/([0-9\/]+)\/tiles`)

	if err != nil {
		return -1, err
	}

	if !re_tiles.MatchString(tiles_uri) {
		return -1, fmt.Errorf("Invalid tiles URI")
	}

	m := re_tiles.FindStringSubmatch(tiles_uri)

	str_id := strings.Replace(m[1], "/", "", -1)

	return strconv.ParseInt(str_id, 10, 64)
}
