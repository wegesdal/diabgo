package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/font/basicfont"

	_ "image/png"
)

const (
	windowWidth  = 1280
	windowHeight = 900
	tileSize     = 64
)

var (
	win          *pixelgl.Window
	actors       []*actor
	characters   []*character
	camPos       = pixel.ZV
	camSpeed     = 500.0
	camZoom      = 1.0
	camZoomSpeed = 1.2
	frame        = 0
	cameraMoved  = true
)

func run() {

	var err error

	levelData, endOfTheRoad := generateMap()

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

	// SPRITESHEETS
	pic, err := loadPicture("dawncastle.png")
	if err != nil {
		panic(err)
	}
	psheet, err := loadPicture("hacker.png")
	if err != nil {
		panic(err)
	}
	tsheet, err := loadPicture("terminal.png")
	if err != nil {
		panic(err)
	}

	// TODO: player respawn on death, tower prototype, potion bar, skill bar, broken bridges (regen levels until path is len > 0)
	// evolutionary approach to enemy composition
	// knights, skeletons, skeleton mages, skeleton warriors
	// track how long they survive and award fitness for completing the level
	// lemmings like terrain alteration / abilities (freezing the water to make a bridge)
	// fix enthrall
	// loot pickups grant abilities (spellbooks)
	// artifacts
	// inventory

	// load player sprites
	var (
		batch         = pixel.NewBatch(&pixel.TrianglesData{}, pic)
		animbatch     = pixel.NewBatch(&pixel.TrianglesData{}, psheet)
		doodadbatch   = pixel.NewBatch(&pixel.TrianglesData{}, pic)
		widgetbatch   = pixel.NewBatch(&pixel.TrianglesData{}, tsheet)
		tiles         = generateTiles(pic)
		player_anim   = generateActorSprites(psheet, 5, 256)
		player_spawn  = findOpenNode(levelData)
		act           = spawn_actor(player_spawn.x, player_spawn.y, "player", player_anim)
		player        = spawn_character(act)
		creep_anim    = generateActorSprites(psheet, 5, 256)
		frames        = 0
		ticks         = 0.0
		second        = time.Tick(time.Second)
		terminal_anim = generateActorSprites(tsheet, 1, 128)
		text_atlas    = text.NewAtlas(basicfont.Face7x13, text.ASCII)
		txt           = text.New(pixel.V(0, 0), text_atlas)
		input         = ""
		last          = time.Now()
	)

	player.maxhp = 80
	player.hp = 80
	player.actor.faction = friendly
	player.prange = 8000.0
	player.arange = 2000.0

	actors = append(actors, act)
	characters = append(characters, player)
	actors = append(actors, spawn_actor(10, 10, "terminal", terminal_anim))

	for !win.Closed() {

		dt := time.Since(last).Seconds()
		imd := imdraw.New(nil)

		ticks += dt

		for _, c := range characters {
			c.actor.coord = pixel.Lerp(c.actor.coord, cartesianToIso(pixel.Vec{X: float64(c.actor.x), Y: float64(c.actor.y)}), dt*4.0)
		}

		if ticks > 0.05 {

			// TODO: APPLY STATUS EFFECTS
			// CHECK STATUS AND APPLY STATE

			characterStateMachine(characters, levelData)
			terminalStateMachine(actors)

			// REMOVE DEAD CHARACTERS AND ADVANCE FRAME

			actors, characters = removeDeadCharacters(actors, characters)

			ticks = 0
		}

		batchUpdate(batch, animbatch, doodadbatch, widgetbatch, txt, actors, dt, levelData, tiles, imd)

		last = time.Now()

		cam := pixel.IM.Scaled(camPos, camZoom).Moved(win.Bounds().Center().Sub(camPos))
		win.SetMatrix(cam)

		if win.JustPressed(pixelgl.Key1) {
			var (
				raw    = isoToCartesian(cam.Unproject(win.MousePosition()))
				coordX = int(raw.X + 1)
				coordY = int(raw.Y + 1)
			)
			if coordX < len(levelData[0]) && coordY < len(levelData[0]) && coordX >= 0 && coordY >= 0 {
				if levelData[0][coordX][coordY] == 0 {
					levelData[0][coordX][coordY] = 4
				} else {
					levelData[0][coordX][coordY] = 0
				}

			}
		}

		if win.JustPressed(pixelgl.Key2) {
			var (
				raw    = isoToCartesian(cam.Unproject(win.MousePosition()))
				coordX = int(raw.X + 1)
				coordY = int(raw.Y + 1)
			)

			if coordX < len(levelData[1]) && coordY < len(levelData[1]) && coordX >= 0 && coordY >= 0 {
				if levelData[1][coordX][coordY] == 0 {
					levelData[1][coordX][coordY] = 1
				} else {
					levelData[1][coordX][coordY] = 0
				}
			}
		}

		if win.JustPressed(pixelgl.Key3) {
			var (
				raw    = isoToCartesian(cam.Unproject(win.MousePosition()))
				coordX = int(raw.X + 1)
				coordY = int(raw.Y + 1)
				act    = spawn_actor(coordX, coordY, "creep", creep_anim)
				c      = spawn_character(act)
			)
			c.dest = endOfTheRoad
			c.actor.faction = hostile
			c.prange = 8000.0
			c.arange = 2000.0
			actors = append(actors, act)
			characters = append(characters, c)
		}

		if win.JustPressed(pixelgl.Key4) {
			for _, c := range characters {
				diff := pixel.Vec.Sub(c.actor.coord, cam.Unproject(win.MousePosition()))
				clickRadius := 300.0
				if diff.X*diff.X+diff.Y*diff.Y < clickRadius {
					c.actor.effects["charmed"] = struct{}{}
					break
				}
			}
		}

		if win.Pressed(pixelgl.MouseButtonLeft) {
			player.actor.state = walk
			var (
				raw    = isoToCartesian(cam.Unproject(win.MousePosition()))
				coordX = int(raw.X + 1)
				coordY = int(raw.Y + 1)
			)

			if coordX < len(levelData[0]) && coordY < len(levelData[0]) && coordX >= 0 && coordY >= 0 {
				player.dest = &node{x: coordX, y: coordY}
			}
		}

		camPos = pixel.Lerp(camPos, player.actor.coord, dt)

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

		// CONFIGURE TERMINAL

		input = handleTerminalInput(player, txt, input)

		win.Clear(pixel.RGBA{R: 0, G: 0, B: 0, A: 0})

		batch.Draw(win)
		drawHealthPlates(characters, imd)
		imd.Draw(win)
		widgetbatch.Draw(win)
		animbatch.Draw(win)
		doodadbatch.Draw(win)

		input = renderTerminalText(player, txt, input)

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

func batchUpdate(batch *pixel.Batch, animbatch *pixel.Batch, doodadbatch *pixel.Batch, widgetbatch *pixel.Batch, txt *text.Text, actors []*actor, dt float64, levelData [2][32][32]uint, tiles [2][]*pixel.Sprite, imd *imdraw.IMDraw) {

	batch.Clear()
	widgetbatch.Clear()
	animbatch.Clear()
	doodadbatch.Clear()

	// only draw tiles close to player
	var player *actor
	for _, a := range actors {
		if a.name == "player" {
			player = a
			break
		}
	}

	if player != nil {

		for x := Min(player.x+16, len(levelData[0])-1); x >= Max(player.x-16, 0); x-- {
			for y := Min(player.y+16, len(levelData[0])-1); y >= Max(player.y-16, 0); y-- {

				isoCoords := cartesianToIso(pixel.V(float64(x), float64(y)))
				mat := pixel.IM.Moved(isoCoords)

				// base map layer
				tiles[0][levelData[0][x][y]].Draw(batch, mat)

				// draw doodads
				if levelData[1][x][y] > 0 {
					tiles[1][levelData[1][x][y]].Draw(doodadbatch, mat)
				}

				// draw all actors (actor animations update every frame)
				for _, a := range actors {
					startingFrame := 0
					// half_length := len(a.anims[a.state]) / 2
					pmat := pixel.IM
					i := isoToCartesian(a.coord)
					// draw actors
					offset := 0.2
					if x == int(i.X+offset) && y == int(i.Y+offset) {
						// DRAW CHARACTER
						// the length of anims tells you if this is a character or item
						// characters will have an anims length of 5
						// widgets will have an anims length of 1
						startingFrame = a.direction * 10

						if len(a.anims) == 5 {
							pmat = pmat.Moved(pixel.Vec.Add(a.coord, pixel.Vec{X: 0, Y: 60}))
							// adjusting Y to account for tall player model
							a.anims[a.state][(a.frame+startingFrame)].Draw(animbatch, pmat)
						} else {

							// DRAW WIDGETS
							pmat = pmat.Moved(pixel.Vec.Add(a.coord, pixel.Vec{X: 25, Y: 80}))
							a.anims[4][(a.frame+startingFrame)].Draw(widgetbatch, pmat)

							isoSquare(a.coord, 3, imd)
						}
					}
				}
			}
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	pixelgl.Run(run)
}
