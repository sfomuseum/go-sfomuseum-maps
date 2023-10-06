package maps

import (
	"fmt"

	"github.com/tidwall/gjson"
)

func DeriveZoomLevels(body []byte) (int, int, error) {

	min_rsp := gjson.GetBytes(body, "properties.mz:min_zoom")

	if !min_rsp.Exists() {
		return 0, 0, fmt.Errorf("Missing mz:min_zoom")
	}

	max_rsp := gjson.GetBytes(body, "properties.mz:max_zoom")

	if !max_rsp.Exists() {
		return 0, 0, fmt.Errorf("Missing mz:max_zoom")
	}

	min_zoom := int(min_rsp.Int())
	max_zoom := int(max_rsp.Int())

	return min_zoom, max_zoom, nil
}
