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

const CRS_PREFIX string = "#CRS: "
const FIELDNAMES string = "mapX,mapY,pixelX,pixelY,enable,dX,dY,residual"

type GroundControlPoints struct {
	CRS    string                `json:"crs"`
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

	crs := ""
	points := make([]*GroundControlPoint, 0)

	buf := bufio.NewReader(r)

	first_ln, err := buf.ReadString('\n')

	if err != nil {
		return nil, err
	}

	offset := int64(0)

	if strings.HasPrefix(first_ln, CRS_PREFIX) {
		crs = strings.Replace(first_ln, CRS_PREFIX, " ", 1)
		crs = strings.TrimSpace(crs)
		offset = int64(len([]byte(first_ln)))
	}

	_, err = r.Seek(int64(offset), 0)

	if err != nil {
		return nil, err
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
		CRS:    crs,
		Points: points,
	}

	return gcp, nil
}

func Marshal(ctx context.Context, gcp *GroundControlPoints, wr io.Writer) error {

	if gcp.CRS != "" {
		crs_ln := fmt.Sprintf("%s%s\n", CRS_PREFIX, gcp.CRS)
		wr.Write([]byte(crs_ln))
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
