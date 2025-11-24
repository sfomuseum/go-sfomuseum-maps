package export

import (
	"context"

	"github.com/sfomuseum/go-sfomuseum-export/v3/properties"
)

func PrepareFeature(ctx context.Context, feature []byte) ([]byte, error) {

	var err error

	feature, err = properties.EnsurePlacetype(feature)

	if err != nil {
		return nil, err
	}

	feature, err = properties.EnsureIsSFO(feature)

	if err != nil {
		return nil, err
	}

	feature, err = properties.EnsureSFOLevel(feature)

	if err != nil {
		return nil, err
	}

	feature, err = properties.EnsureWOFDepicts(feature)

	if err != nil {
		return nil, err
	}

	return feature, nil
}
