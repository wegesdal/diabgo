package main

import (
	"fmt"
	"math"
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

	//https://www.youtube.com/watch?v=Fs13eG36VR8
	//https://opengameart.org/content/60-terrible-character-portraits

	var err error

	cfg := pixelgl.WindowConfig{
		Title:                  "Diabgo",
		Bounds:                 pixel.R(0, 0, windowWidth, windowHeight),
		VSync:                  true,
		TransparentFramebuffer: false,
		Undecorated:            false,
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
	psheet, err := loadPicture("hacker8.png")
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
		levelData, endOfTheRoad = generateMap()
		batch                   = pixel.NewBatch(&pixel.TrianglesData{}, pic)
		animbatch               = pixel.NewBatch(&pixel.TrianglesData{}, psheet)
		doodadbatch             = pixel.NewBatch(&pixel.TrianglesData{}, pic)
		widgetbatch             = pixel.NewBatch(&pixel.TrianglesData{}, tsheet)
		tiles                   = generateTiles(pic)
		player_anim             = generateCharacterSprites(psheet, 256)
		player_spawn            = findOpenNode(levelData)
		act                     = spawn_actor(player_spawn.x, player_spawn.y, "player", player_anim)
		player                  = spawn_character(act)
		creep_anim              = generateCharacterSprites(psheet, 256)
		frames                  = 0
		ticks                   = 0.0
		second                  = time.Tick(time.Second)
		terminal_anim           = generateActorSprites(tsheet, 1, 128)
		text_atlas              = text.NewAtlas(basicfont.Face7x13, text.ASCII)
		term_txt                = text.New(pixel.V(0, 0), text_atlas)
		input                   = ""
		last                    = time.Now()
	)

	player.maxhp = 40
	player.hp = 40
	player.actor.faction = friendly
	player.prange = 0.0
	player.arange = 2000.0

	actors = append(actors, act)
	characters = append(characters, player)
	actors = append(actors, spawn_actor(10, 10, "terminal", terminal_anim))
	//	refreshVisibility(levelData[0], &node{x: player.actor.x, y: player.actor.y})

	for !win.Closed() {

		dt := time.Since(last).Seconds()
		imd := imdraw.New(nil)

		ticks += dt

		for _, c := range characters {
			c.actor.coord = pixel.Lerp(c.actor.coord, cartesianToIso(pixel.Vec{X: float64(c.actor.x), Y: float64(c.actor.y)}), dt*6.0)
		}

		if ticks > 0.03 {
			characterStateMachine(characters, levelData)
			terminalStateMachine(actors)
			actors, characters = removeDeadCharacters(actors, characters)
			ticks = 0
		}

		clearVisibility(levelData[0])
		compute_fov(vec{x: player.actor.x, y: player.actor.y}, levelData[0])
		// refreshVisibility(levelData[0], &node{x: player.actor.x, y: player.actor.y})

		batchUpdate(batch, animbatch, doodadbatch, widgetbatch, term_txt, characters, dt, levelData, tiles, imd)

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
				if levelData[0][coordX][coordY].tile == 0 {
					levelData[0][coordX][coordY].tile = 4
					levelData[0][coordX][coordY].walkable = false
				} else {
					levelData[0][coordX][coordY].tile = 0
					levelData[0][coordX][coordY].walkable = true

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
				if levelData[1][coordX][coordY].tile == 0 {
					levelData[1][coordX][coordY].tile = 1

				} else {
					levelData[1][coordX][coordY].tile = 0
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

		// if win.JustPressed(pixelgl.Key4) {
		// 	for _, c := range characters {
		// 		diff := pixel.Vec.Sub(c.actor.coord, cam.Unproject(win.MousePosition()))
		// 		clickRadius := 300.0
		// 		if diff.X*diff.X+diff.Y*diff.Y < clickRadius {
		// 			c.actor.effects["charmed"] = struct{}{}
		// 			break
		// 		}
		// 	}
		// }

		if win.Pressed(pixelgl.MouseButtonLeft) {
			player.actor.state = walk
			player.target = player
			var (
				raw    = isoToCartesian(cam.Unproject(win.MousePosition()))
				coordX = int(raw.X + 1)
				coordY = int(raw.Y + 1)
			)

			for _, c := range characters {
				// offset y so targeting box is above model
				y_offset := 80.0
				diff := pixel.Vec.Add(pixel.Vec.Sub(c.actor.coord, cam.Unproject(win.MousePosition())), pixel.Vec{X: 0, Y: y_offset})
				if math.Abs(diff.X) < 50 && math.Abs(diff.Y) < 100 && c != player {
					player.target = c
					break
				}
			}
			if coordX < len(levelData[0]) && coordY < len(levelData[0]) && coordX >= 0 && coordY >= 0 {
				player.dest = &node{x: coordX, y: coordY}
			}
		}

		camPos = pixel.Lerp(camPos, player.actor.coord, dt)

		// for i := 0; i < len(characters); i++ {
		// 	_, charmed := characters[i].actor.effects["charmed"]
		// 	if charmed {
		// 		// if actor.effects
		// 		imd.Color = pixel.RGBA{R: 1, G: 0, B: 0.5, A: 0.5}
		// 		imd.EndShape = imdraw.SharpEndShape
		// 		imd.Push(player.actor.coord, characters[i].actor.coord)
		// 		imd.Line(1)
		// 	}
		// }

		// CONFIGURE TERMINAL

		input = handleTerminalInput(player, term_txt, input)

		win.Clear(pixel.RGBA{R: 0.045, G: 0.05, B: 0.105, A: 1.0})

		batch.Draw(win)
		drawHealthPlates(characters, imd)
		imd.Draw(win)
		widgetbatch.Draw(win)
		animbatch.Draw(win)
		doodadbatch.Draw(win)

		input = renderTerminalText(player, term_txt, input)

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

func batchUpdate(batch *pixel.Batch, animbatch *pixel.Batch, doodadbatch *pixel.Batch, widgetbatch *pixel.Batch, term_txt *text.Text, characters []*character, dt float64, levelData [2][32][32]*node, tiles [2][]*pixel.Sprite, imd *imdraw.IMDraw) {

	batch.Clear()
	widgetbatch.Clear()
	animbatch.Clear()
	doodadbatch.Clear()

	// only draw tiles close to player
	var player *character
	for _, c := range characters {
		if c.actor.name == "player" {
			player = c
			break
		}
	}

	if player != nil {

		vision := 16
		for x := Min(player.actor.x+vision, len(levelData[0])-1); x >= Max(player.actor.x-vision, 0); x-- {
			for y := Min(player.actor.y+vision, len(levelData[0])-1); y >= Max(player.actor.y-vision, 0); y-- {

				isoCoords := cartesianToIso(pixel.Vec{X: float64(x), Y: float64(y)})
				mat := pixel.IM.Moved(isoCoords)

				if levelData[0][x][y].visible {
					// base map layer
					tiles[0][levelData[0][x][y].tile].Draw(batch, mat)

					// draw doodads
					if levelData[1][x][y].tile > 0 {
						tiles[1][levelData[1][x][y].tile].Draw(doodadbatch, mat)
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

							if len(a.anims) == 6 {
								pmat = pmat.Moved(pixel.Vec.Add(a.coord, pixel.Vec{X: 0, Y: 60}))
								// adjusting Y to account for tall player model
								a.anims[a.state][(a.frame+startingFrame)].Draw(animbatch, pmat)

								targetRect(player.target.actor.coord, imd, player.target.actor.faction)

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
}

func targetRect(coord pixel.Vec, imd *imdraw.IMDraw, faction int) {
	// imd.Color = pixel.RGBA{R: 1.0, G: 0.6, B: 0, A: 0.5}
	height := 200.0
	y_offset := 80.0
	width := 100.0

	imd.Color = factionColor(faction, light)
	imd.Push(pixel.Vec.Add(coord, pixel.Vec{X: -width / 2, Y: y_offset - height/2}))
	imd.Push(pixel.Vec.Add(coord, pixel.Vec{X: -width / 2, Y: y_offset + height/2}))
	imd.Push(pixel.Vec.Add(coord, pixel.Vec{X: width / 2, Y: y_offset + height/2}))
	imd.Push(pixel.Vec.Add(coord, pixel.Vec{X: width / 2, Y: y_offset - height/2}))
	imd.Polygon(1)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	pixelgl.Run(run)
}
