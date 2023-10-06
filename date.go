package maps

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/tidwall/gjson"
)

var re_yyyy = regexp.MustCompile(`^\d{4}(?:\-\d{2}(?:\-\d{2})?)?$`)

func DeriveYearLabel(body []byte) (string, error) {

	edtf_rsp := gjson.GetBytes(body, "properties.edtf:inception")

	if !edtf_rsp.Exists() {
		return "", fmt.Errorf("Missing edtf:inception")
	}

	if !re_yyyy.MatchString(edtf_rsp.String()) {
		return edtf_rsp.String(), nil
	}

	incept_rsp := gjson.GetBytes(body, "properties.date:inception_upper")

	if !incept_rsp.Exists() {
		return "", fmt.Errorf("Missing date:inception_upper")
	}

	incept_str := incept_rsp.String()
	incept_t, err := time.Parse("2006-01-02", incept_str)

	if err != nil {
		return "", err
	}

	y := incept_t.Year()

	return strconv.Itoa(y), nil

}
