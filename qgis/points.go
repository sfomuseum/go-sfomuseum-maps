package qgis

import (
	"context"
	"fmt"
	"io"
)

type GroundControlPoints struct {
	CSR    string                 `json:"csr"`
	Points []*GroundControlPoints `json:"points"`
}

type GroundControlPoint struct {
	MapX     float64 `json:"mapX"`
	MayY     float64 `json:"mapY"`
	PixelX   float64 `json:"pixelX"`
	PixelY   float64 `json:"pixelY"`
	Enable   float64 `json:"enable"`
	DX       float64 `json"dX"`
	DY       float64 `json"dY"`
	Residual float64 `json:"residual"`
}

func UnmarshalGroundControlPoints(ctx context.Context, r io.Reader) (*GroundControlPoints, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (gcp *GroundControlPoints) Marshal(ctx context.Context, wr io.Writer) error {
	return fmt.Errorf("Not implemented")
}
