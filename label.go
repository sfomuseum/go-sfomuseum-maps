package maps

import (
	"fmt"

	"github.com/tidwall/gjson"
)

func DeriveLabel(body []byte) (string, error) {

	uri_rsp := gjson.GetBytes(body, "properties.sfomuseum:uri")

	if !uri_rsp.Exists() {
		return "", fmt.Errorf("Missing sfomuseum:uri")
	}

	label := uri_rsp.String()
	return label, nil
}
