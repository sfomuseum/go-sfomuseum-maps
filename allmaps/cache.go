package allmaps

import (
	"io"
	"net/http"
	"sync"
)

var data_cache = new(sync.Map)

func fetchURL(allmaps_url string) ([]byte, error) {

	v, exists := data_cache.Load(allmaps_url)

	if exists {
		return v.([]byte), nil
	}

	rsp, err := http.Get(allmaps_url)

	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()

	body, err := io.ReadAll(rsp.Body)

	if err != nil {
		return nil, err
	}

	data_cache.Store(allmaps_url, body)

	return body, nil
}
