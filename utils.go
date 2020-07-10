package main

import (
	"image"
	"math/rand"
	"os"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
)

func cartesianToIso(pt pixel.Vec) pixel.Vec {
	return pixel.V((pt.X-pt.Y)*(tileSize/2), (pt.X+pt.Y)*(tileSize/4))
}

func isoToCartesian(pt pixel.Vec) pixel.Vec {
	x := pt.X*(2.0/tileSize) + pt.Y*(4/tileSize)
	y := ((pt.Y * 4.0 / tileSize) - x) / 2
	return pixel.V(x+y, y)
}

func Max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func Min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func isoSquare(centerXY pixel.Vec, size int, imd *imdraw.IMDraw) {
	imd.Color = pixel.RGBA{R: 1.0, G: 0.6, B: 0, A: 1.0}
	hs := float64(size / 2)
	y_offset := 4.0
	centerXY = pixel.Vec.Add(centerXY, pixel.Vec{X: 0, Y: y_offset})
	imd.Push(pixel.Vec.Add(centerXY, cartesianToIso(pixel.Vec{X: -hs, Y: -hs})))
	imd.Push(pixel.Vec.Add(centerXY, cartesianToIso(pixel.Vec{X: -hs, Y: hs})))
	imd.Push(pixel.Vec.Add(centerXY, cartesianToIso(pixel.Vec{X: hs, Y: hs})))
	imd.Push(pixel.Vec.Add(centerXY, cartesianToIso(pixel.Vec{X: hs, Y: -hs})))
	imd.Polygon(1)
}

func findOpenNode(levelData [2][32][32]*node) *node {
	x := rand.Intn(31)
	y := rand.Intn(31)
	for levelData[0][x][y].tile != 1 {
		x = rand.Intn(31)
		y = rand.Intn(31)
	}
	return &node{x: x, y: y}
}

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}
