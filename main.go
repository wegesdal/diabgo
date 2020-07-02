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
	path      []*node
	direction cardinal
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

	tiles = append(tiles, pixel.NewSprite(pic, pixel.R(0, 128, tileSize, 192)))
	tiles = append(tiles, pixel.NewSprite(pic, pixel.R(0, 448, tileSize, 512)))

	// PLAYER

	// load in player sprites
	playerSprites, err := loadPicture("CitizenSheet.png")
	if err != nil {
		panic(err)
	}
	playerBatch := pixel.NewBatch(&pixel.TrianglesData{}, playerSprites)
	// wizard
	var player = actor{x: 0, y: 0}

	// walking down
	player.frames = append(player.frames, pixel.NewSprite(playerSprites, pixel.R(22*4, 52*3, 22*5, 52*4)))
	player.frames = append(player.frames, pixel.NewSprite(playerSprites, pixel.R(22*5, 52*3, 22*6, 52*4)))
	player.frames = append(player.frames, pixel.NewSprite(playerSprites, pixel.R(22*6, 52*3, 22*7, 52*4)))
	player.frames = append(player.frames, pixel.NewSprite(playerSprites, pixel.R(22*7, 52*3, 22*8, 52*4)))

	// walking up
	player.frames = append(player.frames, pixel.NewSprite(playerSprites, pixel.R(22*8, 52*3, 22*9, 52*4)))
	player.frames = append(player.frames, pixel.NewSprite(playerSprites, pixel.R(22*9, 52*3, 22*10, 52*4)))
	player.frames = append(player.frames, pixel.NewSprite(playerSprites, pixel.R(22*10, 52*3, 22*11, 52*4)))
	player.frames = append(player.frames, pixel.NewSprite(playerSprites, pixel.R(22*11, 52*3, 22*12, 52*4)))

	var (
		frames       = 0
		ticks        = 0.0
		currentFrame = 0
		second       = time.Tick(time.Second)
	)

	last := time.Now()

	updateMap(mapBatch)

	updatePlayer(playerBatch, 0, player)

	for !win.Closed() {

		dt := time.Since(last).Seconds()

		ticks += dt

		if ticks > 0.1 {
			currentFrame++

			if len(player.path) > 0 {

				player.direction.e = player.x - player.path[len(player.path)-1].x
				player.direction.n = player.y - player.path[len(player.path)-1].y

				player.x = player.path[len(player.path)-1].x
				player.y = player.path[len(player.path)-1].y

				player.path = player.path[:len(player.path)-1]
			}

			updatePlayer(playerBatch, currentFrame%4, player)
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
			updateMap(mapBatch)
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
		mapBatch.Draw(win)
		playerBatch.Draw(win)
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

func updatePlayer(batch *pixel.Batch, frame int, player actor) {
	batch.Clear()
	isoCoords := cartesianToIso(pixel.V(float64(player.x), float64(player.y)))
	mat := pixel.IM
	if player.direction.n < 0 || player.direction.e > 0 {
		mat = mat.ScaledXY(pixel.ZV, pixel.V(float64(player.direction.n+-1*player.direction.e), 1))
	}
	startingFrame := 0
	if player.direction.n+player.direction.e < 0 {
		startingFrame = 4
	}

	mat = mat.Moved(isoCoords)

	player.frames[frame+startingFrame].Draw(batch, mat)
}

func updateMap(batch *pixel.Batch) {

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
