package qgis

import (
	"bufio"
	"context"
	"fmt"
	"github.com/sfomuseum/go-csvdict"
	"io"
	"strconv"
	"strings"
)

const CSR_PREFIX string = "#CSR: "
const FIELDNAMES string = "mapX,mapY,pixelX,pixelY,enable,dX,dY,residual"

type GroundControlPoints struct {
	CSR    string                `json:"csr"`
	Points []*GroundControlPoint `json:"points"`
}

type GroundControlPoint struct {
	MapX     float64 `json:"mapX"`
	MapY     float64 `json:"mapY"`
	PixelX   float64 `json:"pixelX"`
	PixelY   float64 `json:"pixelY"`
	Enable   float64 `json:"enable"`
	DX       float64 `json"dX"`
	DY       float64 `json"dY"`
	Residual float64 `json:"residual"`
}

func UnmarshalGroundControlPoints(ctx context.Context, r io.ReadSeeker) (*GroundControlPoints, error) {

	csr := ""
	points := make([]*GroundControlPoint, 0)

	b := bufio.NewReader(r)
	first_ln, err := b.ReadString('\n')

	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(first_ln, CSR_PREFIX) {
		csr = strings.Replace(first_ln, CSR_PREFIX, " ", 1)
	} else {

		_, err = r.Seek(0, 0)

		if err != nil {
			return nil, err
		}
	}

	csv_r, err := csvdict.NewReader(r)

	if err != nil {
		return nil, err
	}

	for {
		row, err := csv_r.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		map_x, err := strconv.ParseFloat(row["mapX"], 64)

		if err != nil {
			return nil, err
		}

		map_y, err := strconv.ParseFloat(row["mapY"], 64)

		if err != nil {
			return nil, err
		}

		pixel_x, err := strconv.ParseFloat(row["pixelX"], 64)

		if err != nil {
			return nil, err
		}

		pixel_y, err := strconv.ParseFloat(row["pixelY"], 64)

		if err != nil {
			return nil, err
		}

		enable, err := strconv.ParseFloat(row["enable"], 64)

		if err != nil {
			return nil, err
		}

		d_x, err := strconv.ParseFloat(row["dX"], 64)

		if err != nil {
			return nil, err
		}

		d_y, err := strconv.ParseFloat(row["dY"], 64)

		if err != nil {
			return nil, err
		}

		residual, err := strconv.ParseFloat(row["residual"], 64)

		if err != nil {
			return nil, err
		}

		pt := &GroundControlPoint{
			MapX:     map_x,
			MapY:     map_y,
			PixelX:   pixel_x,
			PixelY:   pixel_y,
			Enable:   enable,
			DX:       d_x,
			DY:       d_y,
			Residual: residual,
		}

		points = append(points, pt)
	}

	gcp := &GroundControlPoints{
		CSR:    csr,
		Points: points,
	}

	return gcp, nil
}

func (gcp *GroundControlPoints) Marshal(ctx context.Context, wr io.Writer) error {

	if gcp.CSR != "" {
		csr_ln := fmt.Sprintf("%s%s\n", CSR_PREFIX, gcp.CSR)
		wr.Write([]byte(csr_ln))
	}

	csv_wr, err := csvdict.NewWriter(wr, strings.Split(FIELDNAMES, ","))

	if err != nil {
		return err
	}

	csv_wr.WriteHeader()

	for _, pt := range gcp.Points {

		row := map[string]string{
			"mapX":     strconv.FormatFloat(pt.MapX, 'f', -1, 64),
			"mapY":     strconv.FormatFloat(pt.MapY, 'f', -1, 64),
			"pixelX":   strconv.FormatFloat(pt.PixelX, 'f', -1, 64),
			"pixelY":   strconv.FormatFloat(pt.PixelY, 'f', -1, 64),
			"enable":   strconv.FormatFloat(pt.Enable, 'f', 1, 64),
			"dX":       strconv.FormatFloat(pt.DX, 'f', -1, 64),
			"dY":       strconv.FormatFloat(pt.DY, 'f', -1, 64),
			"residual": strconv.FormatFloat(pt.Residual, 'f', -1, 64),
		}

		csv_wr.WriteRow(row)
	}

	return nil
}
