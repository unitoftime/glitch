package assets

import (
	"embed"
	"encoding/json"
	"image"
	"image/draw"
	"io"
)

//go:embed *.png *.json
var FS embed.FS

func LoadImage(path string) (*image.NRGBA, error) {
	file, err := FS.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	nrgba := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(nrgba, nrgba.Bounds(), img, bounds.Min, draw.Src)
	return nrgba, nil
}

func LoadJson(path string, v any) error {
	file, err := FS.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	rawData, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(rawData, v)
}
