package allmaps

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	
	"github.com/tidwall/gjson"
)

func DeriveImageId(allmaps_url string) (string, error) {
	
	u, err := url.Parse(allmaps_url)

	if err != nil {
		return "", err
	}

	if u.Host != "annotations.allmaps.org" {
		return "", fmt.Errorf("Invalid host")
	}

	// https://annotations.allmaps.org/images/c157f0e8c25aa123.geojson
	
	if strings.HasPrefix(u.Path, "/images/"){

		fname := filepath.Base(u.Path)
		ext := filepath.Ext(u.Path)
		
		switch ext {
		case ".geojson":
			image_id := strings.Replace(fname, ext, "", 1)
			return image_id, nil
		default:
			return "", fmt.Errorf("Unsupported URL")
		}		
	}

	// https://annotations.allmaps.org/maps/a0c0c652e49f4596
	
	if strings.HasPrefix(u.Path, "/maps/"){
		
		rsp, err := http.Get(allmaps_url)

		if err != nil {
			return "", err
		}
		
		defer rsp.Body.Close()
		
		body, err := io.ReadAll(rsp.Body)
		
		if err != nil {
			return "", fmt.Errorf("Failed to read body, %w", err)
		}
		
		id_rsp := gjson.GetBytes(body, "body._allmaps.image.id")
		
		if !id_rsp.Exists(){
			return "", fmt.Errorf("Body missing image ID")
		}

		image_id := filepath.Base(id_rsp.String())
		return image_id, nil
	}

	return "", fmt.Errorf("Unsupported URL")
}
