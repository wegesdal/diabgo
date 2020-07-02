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

type cardinal struct {
	n int
	e int
}

type actor struct {
	x         int
	y         int
	frames    []*pixel.Sprite
	frame     int
	path      []*node
	direction cardinal
}

type vec2 struct {
	x float64
	y float64
}

func lerp(start pixel.Vec, end pixel.Vec, percent float64) pixel.Vec {
	x := start.X + percent*(end.X-start.X)
	y := start.Y + percent*(end.Y-start.Y)
	return (pixel.Vec{X: x, Y: y})
}

const (
	windowWidth  = 800
	windowHeight = 600
	// sprite tiles are squared, 64x64 size
	tileSize = 64
	f        = 0 // floor identifier
	w        = 1 // wall identifier
)

var levelData = [32][32]uint{}

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
		VSync:  false,
	}
	win, err = pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// MAP

	pic, err := loadPicture("dawncastle.png")
	if err != nil {
		panic(err)
	}

	mapBatch := pixel.NewBatch(&pixel.TrianglesData{}, pic)

	tiles = append(tiles, pixel.NewSprite(pic, pixel.R(0, 64, tileSize, 128)))
	tiles = append(tiles, pixel.NewSprite(pic, pixel.R(0, 0, tileSize, 64)))

	// PLAYER

	// load in player sprites
	// playerSprites, err := loadPicture("dawncastle.png")
	// if err != nil {
	// 	panic(err)
	// }
	// playerBatch := pixel.NewBatch(&pixel.TrianglesData{}, playerSprites)
	// wizard
	var player = actor{x: 0, y: 0, frame: 0}

	var min_Y float64 = 52 * 3
	var max_Y float64 = 52*4 - 10
	var min_X float64 = 466.0

	// walking down
	player.frames = append(player.frames, pixel.NewSprite(pic, pixel.R(min_X+22*0, min_Y, min_X+22*1, max_Y)))
	player.frames = append(player.frames, pixel.NewSprite(pic, pixel.R(min_X+22*1, min_Y, min_X+22*2, max_Y)))
	player.frames = append(player.frames, pixel.NewSprite(pic, pixel.R(min_X+22*2, min_Y, min_X+22*3, max_Y)))
	player.frames = append(player.frames, pixel.NewSprite(pic, pixel.R(min_X+22*3, min_Y, min_X+22*4, max_Y)))

	// walking up
	player.frames = append(player.frames, pixel.NewSprite(pic, pixel.R(min_X+22*4, min_Y, min_X+22*5, max_Y)))
	player.frames = append(player.frames, pixel.NewSprite(pic, pixel.R(min_X+22*5, min_Y, min_X+22*6, max_Y)))
	player.frames = append(player.frames, pixel.NewSprite(pic, pixel.R(min_X+22*6, min_Y, min_X+22*7, max_Y)))
	player.frames = append(player.frames, pixel.NewSprite(pic, pixel.R(min_X+22*7, min_Y, min_X+22*8, max_Y)))

	var (
		frames = 0
		ticks  = 0.0
		second = time.Tick(time.Second)
	)

	last := time.Now()

	updateMap(mapBatch, player)

	//updatePlayer(mapBatch, player)

	for !win.Closed() {

		dt := time.Since(last).Seconds()

		ticks += dt

		if ticks > 0.1 {

			player.frame = (player.frame + 1) % 4

			if len(player.path) > 0 {

				player.direction.e = player.x - player.path[len(player.path)-1].x
				player.direction.n = player.y - player.path[len(player.path)-1].y

				player.x = player.path[len(player.path)-1].x
				player.y = player.path[len(player.path)-1].y

				player.path = player.path[:len(player.path)-1]
			}

			//updatePlayer(mapBatch, player)
			updateMap(mapBatch, player)
			ticks = 0
		}

		last = time.Now()

		cam := pixel.IM.Scaled(camPos, camZoom).Moved(win.Bounds().Center().Sub(camPos))
		win.SetMatrix(cam)

		if win.JustPressed(pixelgl.Key1) {
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
			updateMap(mapBatch, player)
		}

		if win.JustPressed(pixelgl.MouseButtonLeft) {
			var raw = isoToCartesian(cam.Unproject(win.MousePosition()))

			var coordX = int(raw.X + 1)
			var coordY = int(raw.Y + 1)

			if coordX < len(levelData) && coordY < len(levelData[0]) && coordX >= 0 && coordY >= 0 {
				player.path = Astar(&node{x: player.x, y: player.y}, &node{x: coordX, y: coordY}, levelData)
				//print(player.path)
			}
		}

		camPos = lerp(camPos, cartesianToIso(pixel.Vec{X: float64(player.x), Y: float64(player.y)}), dt)

		// manual camera

		// if win.Pressed(pixelgl.KeyLeft) {
		// 	camPos.X -= camSpeed * dt
		// }
		// if win.Pressed(pixelgl.KeyRight) {
		// 	camPos.X += camSpeed * dt
		// }
		// if win.Pressed(pixelgl.KeyDown) {
		// 	camPos.Y -= camSpeed * dt
		// }
		// if win.Pressed(pixelgl.KeyUp) {
		// 	camPos.Y += camSpeed * dt
		// }

		// camZoom *= math.Pow(camZoomSpeed, win.MouseScroll().Y)

		win.Clear(colornames.Darkmagenta)
		mapBatch.Draw(win)
		// playerBatch.Draw(win)
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

// TODO: combine the player sprite and map sprite sheet to batch all at once

func updateMap(batch *pixel.Batch, player actor) {

	batch.Clear()

	for x := len(levelData) - 1; x >= 0; x-- {
		for y := len(levelData[x]) - 1; y >= 0; y-- {

			pmat := pixel.IM
			startingFrame := 0

			isoCoords := cartesianToIso(pixel.V(float64(x), float64(y)))

			mat := pixel.IM.Moved(isoCoords)
			tiles[levelData[x][y]].Draw(batch, mat)

			if x == player.x && y == player.y {

				if player.direction.n < 0 || player.direction.e > 0 {
					pmat = pmat.ScaledXY(pixel.ZV, pixel.V(float64(player.direction.n+-1*player.direction.e), 1))
				}

				if player.direction.n+player.direction.e < 0 {
					startingFrame = 4
				}

				pmat = pmat.Moved(isoCoords)
				player.frames[player.frame+startingFrame].Draw(batch, pmat)
			}
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
