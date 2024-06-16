package assets

import (
	"embed"
	"image"
	"image/draw"
)

//go:embed *.png
var FS embed.FS

func LoadImage(path string) (*image.NRGBA, error) {
	file, err := FS.Open(path)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	nrgba := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(nrgba, nrgba.Bounds(), img, bounds.Min, draw.Src)
	return nrgba, nil
}
