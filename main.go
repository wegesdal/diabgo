package main

import (
	"fmt"
	"image"
	"math/rand"
	"os"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"

	_ "image/png"
)

const (
	windowWidth  = 1280
	windowHeight = 900
	tileSize     = 64
)

var win *pixelgl.Window

var actors []*actor
var characters []*character

var (
	camPos       = pixel.ZV
	camSpeed     = 500.0
	camZoom      = 1.0
	camZoomSpeed = 1.2
	frame        = 0
	cameraMoved  = true
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

func findOpenNode(levelData [2][32][32]uint) *node {
	x := rand.Intn(31)
	y := rand.Intn(31)
	for levelData[0][x][y] != 1 {
		x = rand.Intn(31)
		y = rand.Intn(31)
	}
	return &node{x: x, y: y}
}

func run() {

	var err error
	endOfTheRoad := &node{x: rand.Intn(31), y: 31}
	levelData := generateMap(endOfTheRoad)

	cfg := pixelgl.WindowConfig{
		Title:                  "Diabgo",
		Bounds:                 pixel.R(0, 0, windowWidth, windowHeight),
		VSync:                  true,
		TransparentFramebuffer: true,
		Undecorated:            true,
	}
	win, err = pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// SPRITESHEET
	pic, err := loadPicture("dawncastle.png")
	if err != nil {
		panic(err)
	}

	// SPRITESHEET
	psheet, err := loadPicture("test2.png")
	if err != nil {
		panic(err)
	}

	batch := pixel.NewBatch(&pixel.TrianglesData{}, pic)
	animbatch := pixel.NewBatch(&pixel.TrianglesData{}, psheet)
	doodadbatch := pixel.NewBatch(&pixel.TrianglesData{}, pic)
	tiles := generateTiles(pic)

	// TODO: player respawn on death, tower prototype, potion bar, skill bar, broken bridges (regen levels until path is len > 0)
	// evolutionary approach to enemy composition
	// knights, skeletons, skeleton mages, skeleton warriors
	// track how long they survive and award fitness for completing the level
	// lemmings like terrain alteration / abilities (freezing the water to make a bridge)
	// fix enthrall
	// enchantress artwork
	// loot pickups grant abilities (spellbooks)
	// artifacts
	// inventory

	// load player sprites
	var player_anim = generateActorSprites(psheet)

	player_spawn := findOpenNode(levelData)
	var player = spawn_character(spawn_actor(player_spawn.x, player_spawn.y, "player", player_anim))
	player.maxhp = 80
	player.hp = 80
	player.actor.faction = friendly
	player.prange = 8000.0
	player.arange = 2000.0

	characters = append(characters, player)

	var creep_anim = generateActorSprites(psheet)

	var (
		frames = 0
		ticks  = 0.0
		second = time.Tick(time.Second)
	)

	last := time.Now()

	for !win.Closed() {

		dt := time.Since(last).Seconds()

		ticks += dt

		for _, c := range characters {
			c.actor.coord = pixel.Lerp(c.actor.coord, cartesianToIso(pixel.Vec{X: float64(c.actor.x), Y: float64(c.actor.y)}), dt*4.0)
		}

		if ticks > 0.05 {

			// TODO: APPLY STATUS EFFECTS
			// CHECK STATUS AND APPLY STATE

			characterStateMachine(characters, levelData)

			// REMOVE DEAD ACTORS AND ADVANCE FRAME
			for i, c := range characters {

				c.actor.frame = (c.actor.frame + 1) % 10
				// KILL CREEPS WHO REACH END OF THE ROAD
				// TODO: ADJUST SCORE
				if c.actor.name == "creep" && c.actor.x == c.dest.x && c.actor.y == c.dest.y && c.actor.state != dead {
					c.actor.frame = 0
					c.actor.state = dead
				}
				if c.actor.state == dead && c.actor.frame == 9 {
					characters[i] = characters[len(characters)-1]
					characters[len(characters)-1] = nil
					characters = characters[:len(characters)-1]
					break
				}
			}
			ticks = 0
		}

		text_atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
		txt := text.New(pixel.Vec.Sub(camPos, pixel.Vec{X: 100.0, Y: -100.0}), text_atlas)

		batchUpdate(batch, animbatch, doodadbatch, txt, actors, dt, levelData, tiles)

		last = time.Now()

		cam := pixel.IM.Scaled(camPos, camZoom).Moved(win.Bounds().Center().Sub(camPos))
		win.SetMatrix(cam)

		if win.JustPressed(pixelgl.Key1) {
			var raw = isoToCartesian(cam.Unproject(win.MousePosition()))

			var coordX = int(raw.X + 1)
			var coordY = int(raw.Y + 1)

			if coordX < len(levelData[0]) && coordY < len(levelData[0]) && coordX >= 0 && coordY >= 0 {
				if levelData[0][coordX][coordY] == 0 {
					levelData[0][coordX][coordY] = 4
				} else {
					levelData[0][coordX][coordY] = 0
				}

			}
		}

		if win.JustPressed(pixelgl.Key2) {
			var raw = isoToCartesian(cam.Unproject(win.MousePosition()))

			var coordX = int(raw.X + 1)
			var coordY = int(raw.Y + 1)

			if coordX < len(levelData[1]) && coordY < len(levelData[1]) && coordX >= 0 && coordY >= 0 {
				if levelData[1][coordX][coordY] == 0 {
					levelData[1][coordX][coordY] = 1
				} else {
					levelData[1][coordX][coordY] = 0
				}
			}
		}

		if win.JustPressed(pixelgl.Key3) {
			var raw = isoToCartesian(cam.Unproject(win.MousePosition()))
			var coordX = int(raw.X + 1)
			var coordY = int(raw.Y + 1)
			c := spawn_character(spawn_actor(coordX, coordY, "creep", creep_anim))
			c.dest = endOfTheRoad
			c.actor.faction = hostile
			c.prange = 8000.0
			c.arange = 2000.0
			characters = append(characters, c)
		}

		if win.JustPressed(pixelgl.Key4) {
			for _, a := range actors {
				diff := pixel.Vec.Sub(a.coord, cam.Unproject(win.MousePosition()))
				clickRadius := 300.0
				if diff.X*diff.X+diff.Y*diff.Y < clickRadius {
					a.effects["charmed"] = struct{}{}
					break
				}
			}
		}

		if win.Pressed(pixelgl.MouseButtonLeft) {
			var raw = isoToCartesian(cam.Unproject(win.MousePosition()))

			var coordX = int(raw.X + 1)
			var coordY = int(raw.Y + 1)

			if coordX < len(levelData[0]) && coordY < len(levelData[0]) && coordX >= 0 && coordY >= 0 {
				player.dest = &node{x: coordX, y: coordY}
			}
		}

		oldPos := camPos
		camPos = pixel.Lerp(camPos, player.actor.coord, dt)
		if oldPos == camPos {
			cameraMoved = false
		} else {
			cameraMoved = true
		}

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

		// ... draw the scene using imd

		imd := imdraw.New(nil)

		for i := 0; i < len(characters); i++ {
			_, charmed := characters[i].actor.effects["charmed"]
			if charmed {
				// if actor.effects
				imd.Color = pixel.RGBA{R: 1, G: 0, B: 0.5, A: 0.5}
				imd.EndShape = imdraw.SharpEndShape
				imd.Push(player.actor.coord, characters[i].actor.coord)
				imd.Line(1)
			}
		}

		win.Clear(pixel.RGBA{R: 0, G: 0, B: 0, A: 0})
		if cameraMoved {
			batch.Draw(win)
		}
		drawHealthPlates(characters, imd)
		imd.Draw(win)

		animbatch.Draw(win)
		doodadbatch.Draw(win)
		txt.Draw(win, pixel.IM.Scaled(txt.Orig, 4))

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

func batchUpdate(batch *pixel.Batch, animbatch *pixel.Batch, doodadbatch *pixel.Batch, txt *text.Text, actors []*actor, dt float64, levelData [2][32][32]uint, tiles [2][]*pixel.Sprite) {

	// TODO: I only need to clear this if the camera moved last tick.
	batch.Clear()
	animbatch.Clear()
	doodadbatch.Clear()

	// not tiles display on top of character atm

	// only draw tiles close to
	var player *character
	for _, c := range characters {
		if c.actor.name == "player" {
			player = c
			break
		}
	}

	if player != nil {

		for x := Min(player.actor.x+16, len(levelData[0])-1); x >= Max(player.actor.x-16, 0); x-- {
			for y := Min(player.actor.y+16, len(levelData[0])-1); y >= Max(player.actor.y-16, 0); y-- {

				isoCoords := cartesianToIso(pixel.V(float64(x), float64(y)))
				mat := pixel.IM.Moved(isoCoords)

				// base map layer
				tiles[0][levelData[0][x][y]].Draw(batch, mat)

				// draw doodads
				if levelData[1][x][y] > 0 {
					tiles[1][levelData[1][x][y]].Draw(doodadbatch, mat)
				}

				for _, c := range characters {
					startingFrame := 0
					// half_length := len(a.anims[a.state]) / 2
					pmat := pixel.IM
					i := isoToCartesian(c.actor.coord)
					// draw actors
					offset := 0.2
					if x == int(i.X+offset) && y == int(i.Y+offset) {

						startingFrame = c.actor.direction * 10
						pmat = pmat.Moved(pixel.Vec.Add(c.actor.coord, pixel.Vec{X: 0, Y: 60}))

						c.actor.anims[c.actor.state][(c.actor.frame+startingFrame)].Draw(animbatch, pmat)
					}
				}
			}
		}
	} else {
		txt.Color = colornames.Lightgreen
		fmt.Fprintln(txt, "You Died")
		// game over
	}
}

func cartesianToIso(pt pixel.Vec) pixel.Vec {
	return pixel.V((pt.X-pt.Y)*(tileSize/2), (pt.X+pt.Y)*(tileSize/4))
}

func isoToCartesian(pt pixel.Vec) pixel.Vec {
	x := pt.X*(2.0/tileSize) + pt.Y*(4/tileSize)
	y := ((pt.Y * 4.0 / tileSize) - x) / 2
	return pixel.V(x+y, y)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	pixelgl.Run(run)
}
