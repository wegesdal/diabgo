package main

import (
	"fmt"
	"image"
	"os"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"

	_ "image/png"
)

const (
	windowWidth  = 800
	windowHeight = 800
	// sprite tiles are squared, 64x64 size
	tileSize = 64
	f        = 0 // floor identifier
	w        = 1 // wall identifier
)

var levelData = [][]uint{
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // This row will be rendered in the lower left part of the screen (closer to the viewer)
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // And this in the upper right
}

var win *pixelgl.Window
var floorTile, wallTile *pixel.Sprite
var tiles []*pixel.Sprite

var (
	camPos       = pixel.ZV
	camSpeed     = 500.0
	camZoom      = 1.0
	camZoomSpeed = 1.2
)

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

func run() {

	var err error

	cfg := pixelgl.WindowConfig{
		Title:  "Diabgo",
		Bounds: pixel.R(0, 0, windowWidth, windowHeight),
		VSync:  true,
	}
	win, err = pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	pic, err := loadPicture("dawncastle.png")
	if err != nil {
		panic(err)
	}

	batch := pixel.NewBatch(&pixel.TrianglesData{}, pic)

	tiles = append(tiles, pixel.NewSprite(pic, pixel.R(0, 128, tileSize, 192)))
	tiles = append(tiles, pixel.NewSprite(pic, pixel.R(0, 448, tileSize, 512)))

	var (
		frames = 0
		second = time.Tick(time.Second)
	)

	last := time.Now()

	updateBatch(batch)

	for !win.Closed() {

		dt := time.Since(last).Seconds()
		last = time.Now()

		cam := pixel.IM.Scaled(camPos, camZoom).Moved(win.Bounds().Center().Sub(camPos))
		win.SetMatrix(cam)

		if win.JustPressed(pixelgl.MouseButtonLeft) {
			var raw = isoToCartesian(cam.Unproject(win.MousePosition()))

			var coordX = int(raw.X + 1)
			var coordY = int(raw.Y + 1)

			if coordX < len(levelData) && coordY < len(levelData[0]) && coordX >= 0 && coordY >= 0 {
				if levelData[coordX][coordY] == 0 {
					levelData[coordX][coordY] = 1
				} else {
					levelData[coordX][coordY] = 0
				}

			}
			updateBatch(batch)
		}

		if win.Pressed(pixelgl.KeyLeft) {
			camPos.X -= camSpeed * dt
		}
		if win.Pressed(pixelgl.KeyRight) {
			camPos.X += camSpeed * dt
		}
		if win.Pressed(pixelgl.KeyDown) {
			camPos.Y -= camSpeed * dt
		}
		if win.Pressed(pixelgl.KeyUp) {
			camPos.Y += camSpeed * dt
		}
		// camZoom *= math.Pow(camZoomSpeed, win.MouseScroll().Y)

		win.Clear(colornames.Darkmagenta)
		batch.Draw(win)
		win.Update()

		frames++

		select {
		case <-second:
			win.SetTitle(fmt.Sprintf("%s | FPS: %d", cfg.Title, frames))
			frames = 0
		default:
		}
	}
}

func updateBatch(batch *pixel.Batch) {

	batch.Clear()
	for x := len(levelData) - 1; x >= 0; x-- {
		for y := len(levelData[x]) - 1; y >= 0; y-- {
			isoCoords := cartesianToIso(pixel.V(float64(x), float64(y)))
			mat := pixel.IM.Moved(isoCoords)
			// Not really needed, just put to show bigger blocks
			// mat = mat.ScaledXY(win.Bounds().Center(), pixel.V(2, 2))
			tiles[levelData[x][y]].Draw(batch, mat)
		}
	}

}

func cartesianToIso(pt pixel.Vec) pixel.Vec {
	return pixel.V((pt.X-pt.Y)*(tileSize/2), (pt.X+pt.Y)*(tileSize/4))
}

func isoToCartesian(pt pixel.Vec) pixel.Vec {
	var xx = pt.X*(2.0/tileSize) + pt.Y*(4/tileSize)
	var yy = ((pt.Y * 4.0 / tileSize) - xx) / 2
	return pixel.V(xx+yy, yy)
}

func main() {
	pixelgl.Run(run)
}
