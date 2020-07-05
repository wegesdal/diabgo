package main

import (
	"fmt"
	"image"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"

	_ "image/png"
)

type cardinal struct {
	n int
	e int
}

const (
	walk = iota
	attack
	dead
	idle
	cast
)

const (
	hostile = iota - 1
	neutral
	friendly
)

type actor struct {
	x         int
	y         int
	name      string
	coord     pixel.Vec
	frames    []*pixel.Sprite
	maxhp     int
	hp        int
	state     int
	faction   int
	movespeed float64
	effects   map[string]struct{}
	prange    float64
	arange    float64
	dest      *node
	direction cardinal
	target    *actor
	anims     map[int][]*pixel.Sprite
}

// EFFECTS NOTES: you can get a pseudo set feature with the following:
// map[string] struct{}
// value, ok := yourmap[key]
// struct{}{}

const (
	windowWidth  = 1280
	windowHeight = 900
	// sprite tiles are squared, 64x64 size
	tileSize = 64
)

var levelData = [32][32]uint{}
var doodadData = [32][32]uint{}

var win *pixelgl.Window
var tiles []*pixel.Sprite
var doodads []*pixel.Sprite

var actors []*actor

var (
	camPos       = pixel.ZV
	camSpeed     = 500.0
	camZoom      = 1.0
	camZoomSpeed = 1.2
	frame        = 0
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

func wall_gen(x int, y int) {
	levelData[x][y] = 4
	r := rand.Intn(6)
	if x < 30 && x > 0 && y < 30 && y > 0 {
		if r < 2 {
			wall_gen(x+1, y)
		} else if r < 4 {
			wall_gen(x, y+1)
		}
	}
}

func spawn_actor(x int, y int, name string, anims map[int][]*pixel.Sprite) *actor {
	var a = actor{x: x, y: y}
	a.name = name
	a.anims = anims
	a.maxhp = 40
	a.hp = 15
	a.dest = &node{x: x, y: y}
	a.coord = cartesianToIso(pixel.V(float64(a.x), float64(a.y)))
	a.effects = map[string]struct{}{}
	a.state = idle
	a.target = &a
	return &a
}

func step_forward(a *actor, path []*node) {
	if len(path) > 0 {
		i := isoToCartesian(a.coord)
		// don't update next block until close
		if math.Pow(i.X-float64(a.x), 2.0)+math.Pow(i.Y-float64(a.y), 2.0) < 1 {
			a.direction.e = a.x - path[len(path)-1].x
			a.direction.n = a.y - path[len(path)-1].y
			a.x = path[len(path)-1].x
			a.y = path[len(path)-1].y
		}
	}
}

func run() {

	var err error

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

	// MAP
	pic, err := loadPicture("dawncastle.png")
	if err != nil {
		panic(err)
	}

	mapBatch := pixel.NewBatch(&pixel.TrianglesData{}, pic)

	// ground
	tiles = append(tiles, pixel.NewSprite(pic, pixel.R(0, 64, tileSize, 128)))

	// road
	tiles = append(tiles, pixel.NewSprite(pic, pixel.R(0, 128, tileSize, 192)))

	// bridge
	tiles = append(tiles, pixel.NewSprite(pic, pixel.R(64, 256, tileSize+64, 256-64)))
	tiles = append(tiles, pixel.NewSprite(pic, pixel.R(0, 256, tileSize, 256-64)))

	// block
	tiles = append(tiles, pixel.NewSprite(pic, pixel.R(0, 0, tileSize, 64)))

	// water
	tiles = append(tiles, pixel.NewSprite(pic, pixel.R(0+64, 128, tileSize+64, 192)))

	// the 0th doodad is nil to allow direct grid assignment (empty array is 0)
	doodads = append(doodads, nil)
	doodads = append(doodads, pixel.NewSprite(pic, pixel.R(0+128, 128-64, 128+128, 192+64)))

	// MAP GENERATION
	// TODO: Break out into its own file

	// make some random trees
	for i := 0; i < rand.Intn(60); i++ {
		doodadData[rand.Intn(31)][rand.Intn(31)] = 1
	}

	// make some walls
	for i := 0; i < 10; i++ {

		start_x := rand.Intn(25) + 4
		start_y := rand.Intn(25) + 4
		wall_gen(start_x, start_y)
	}

	// generate a path
	road_start := &node{x: rand.Intn(31), y: 0}
	road := Astar(road_start, &node{x: rand.Intn(31), y: 31}, levelData)

	// generate a river
	river_start := &node{x: 0, y: rand.Intn(31)}
	river := Astar(river_start, &node{x: 31, y: rand.Intn(31)}, levelData)

	// bake the road onto the array
	for _, node := range road {
		levelData[node.x][node.y] = 1
	}
	// bake the river onto the array
	river = append(river, river_start)

	for _, node := range river {
		if levelData[node.x][node.y] == 1 {
			levelData[node.x][node.y-1] = 2
			levelData[node.x][node.y] = 3
			levelData[node.x][node.y+1] = 3

		} else {
			levelData[node.x][node.y] = 5
			levelData[node.x][node.y+1] = 5

		}
	}

	// PLAYER
	// load in player sprites

	var player_anim = make(map[int][]*pixel.Sprite)

	var min_X float64 = 0.0
	var min_Y float64 = 512 - 192

	// PLAYER ANIMATION FRAMES
	for i := 0; i < 28; i++ {
		if i < 8 {
			player_anim[walk] = append(player_anim[walk], pixel.NewSprite(pic, pixel.R(min_X+64*float64(i), min_Y, min_X+64*float64(i+1), min_Y+64)))
		} else if i < 16 {
			player_anim[attack] = append(player_anim[attack], pixel.NewSprite(pic, pixel.R(min_X+64*float64(i), min_Y, min_X+64*float64(i+1), min_Y+64)))
		} else if i < 18 {
			player_anim[dead] = append(player_anim[dead], pixel.NewSprite(pic, pixel.R(min_X+64*float64(i), min_Y, min_X+64*float64(i+1), min_Y+64)))
		} else if i < 20 {
			player_anim[idle] = append(player_anim[idle], pixel.NewSprite(pic, pixel.R(min_X+64*float64(i), min_Y, min_X+64*float64(i+1), min_Y+64)))
		} else if i < 28 {
			player_anim[cast] = append(player_anim[cast], pixel.NewSprite(pic, pixel.R(min_X+64*float64(i), min_Y, min_X+64*float64(i+1), min_Y+64)))
		}
	}

	var player = spawn_actor(0, 0, "player", player_anim)
	player.maxhp = 80
	player.hp = 80
	player.faction = friendly
	player.prange = 8000.0
	player.arange = 2000.0

	actors = append(actors, player)

	// CREEP ANIMATION FRAMES
	min_Y += 64

	var creep_anim = make(map[int][]*pixel.Sprite)

	// PLAYER ANIMATION FRAMES
	for i := 0; i < 28; i++ {
		if i < 8 {
			creep_anim[walk] = append(creep_anim[walk], pixel.NewSprite(pic, pixel.R(min_X+64*float64(i), min_Y, min_X+64*float64(i+1), min_Y+64)))
		} else if i < 16 {
			creep_anim[attack] = append(creep_anim[attack], pixel.NewSprite(pic, pixel.R(min_X+64*float64(i), min_Y, min_X+64*float64(i+1), min_Y+64)))
		} else if i < 18 {
			creep_anim[dead] = append(creep_anim[dead], pixel.NewSprite(pic, pixel.R(min_X+64*float64(i), min_Y, min_X+64*float64(i+1), min_Y+64)))
		} else if i < 20 {
			creep_anim[idle] = append(creep_anim[idle], pixel.NewSprite(pic, pixel.R(min_X+64*float64(i), min_Y, min_X+64*float64(i+1), min_Y+64)))
		} else if i < 28 {
			creep_anim[cast] = append(creep_anim[cast], pixel.NewSprite(pic, pixel.R(min_X+64*float64(i), min_Y, min_X+64*float64(i+1), min_Y+64)))
		}
	}

	var (
		frames = 0
		ticks  = 0.0
		second = time.Tick(time.Second)
	)

	last := time.Now()

	for !win.Closed() {

		dt := time.Since(last).Seconds()

		ticks += dt

		for _, a := range actors {
			a.coord = pixel.Lerp(a.coord, cartesianToIso(pixel.Vec{X: float64(a.x), Y: float64(a.y)}), dt*4.0)
		}

		if ticks > 0.1 {

			frame = (frame + 1) % 4

			// TODO: APPLY STATUS EFFECTS
			// CHECK STATUS AND APPLY STATE

			actorStateMachine(actors)

			// REMOVE DEAD ACTORS
			for _, a := range actors {
				// KILL CREEPS WHO REACH END OF THE ROAD
				// TODO: ADJUST SCORE
				if (a.name == "creep" && a.x == a.dest.x && a.y == a.dest.y) || a.hp < 1 {
					a.state = dead
				}
				// if a.state == dead {
				// 	actors[i] = actors[len(actors)-1]
				// 	actors[len(actors)-1] = nil
				// 	actors = actors[:len(actors)-1]
				// 	break
				// }
			}
			ticks = 0
		}

		updateMap(mapBatch, actors, dt)

		last = time.Now()

		cam := pixel.IM.Scaled(camPos, camZoom).Moved(win.Bounds().Center().Sub(camPos))
		win.SetMatrix(cam)

		if win.JustPressed(pixelgl.Key1) {
			var raw = isoToCartesian(cam.Unproject(win.MousePosition()))

			var coordX = int(raw.X + 1)
			var coordY = int(raw.Y + 1)

			if coordX < len(levelData) && coordY < len(levelData[0]) && coordX >= 0 && coordY >= 0 {
				if levelData[coordX][coordY] == 0 {
					levelData[coordX][coordY] = 4
				} else {
					levelData[coordX][coordY] = 0
				}

			}
		}

		if win.JustPressed(pixelgl.Key2) {
			var raw = isoToCartesian(cam.Unproject(win.MousePosition()))

			var coordX = int(raw.X + 1)
			var coordY = int(raw.Y + 1)

			if coordX < len(doodadData) && coordY < len(doodadData[0]) && coordX >= 0 && coordY >= 0 {
				if doodadData[coordX][coordY] == 0 {
					doodadData[coordX][coordY] = 1
				} else {
					doodadData[coordX][coordY] = 0
				}
			}
		}

		if win.JustPressed(pixelgl.Key3) {
			var raw = isoToCartesian(cam.Unproject(win.MousePosition()))
			var coordX = int(raw.X + 1)
			var coordY = int(raw.Y + 1)
			c := spawn_actor(coordX, coordY, "creep", creep_anim)
			c.dest = road[0]
			c.faction = hostile
			c.prange = 8000.0
			c.arange = 2000.0
			actors = append(actors, c)
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

			if coordX < len(levelData) && coordY < len(levelData[0]) && coordX >= 0 && coordY >= 0 {
				player.dest = &node{x: coordX, y: coordY}
			}
		}

		camPos = pixel.Lerp(camPos, player.coord, dt)

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

		for i := 0; i < len(actors); i++ {
			_, charmed := actors[i].effects["charmed"]
			if charmed {
				// if actor.effects
				imd.Color = pixel.RGBA{R: 1, G: 0, B: 0.5, A: 0.5}
				imd.EndShape = imdraw.SharpEndShape
				imd.Push(player.coord, actors[i].coord)
				imd.Line(1)
			}
		}

		// DRAW HEALTH PLATES
		drawHealthPlates(actors, imd)

		win.Clear(pixel.RGBA{R: 0, G: 0, B: 0, A: 0})
		mapBatch.Draw(win)
		imd.Draw(win)
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

func updateMap(batch *pixel.Batch, actors []*actor, dt float64) {

	batch.Clear()

	for x := len(levelData) - 1; x >= 0; x-- {
		for y := len(levelData[x]) - 1; y >= 0; y-- {

			isoCoords := cartesianToIso(pixel.V(float64(x), float64(y)))
			mat := pixel.IM.Moved(isoCoords)
			tiles[levelData[x][y]].Draw(batch, mat)

			// draw doodads
			if doodadData[x][y] > 0 {
				doodads[doodadData[x][y]].Draw(batch, mat)
			}

			for _, a := range actors {
				startingFrame := 0
				half_length := len(a.anims[a.state]) / 2
				pmat := pixel.IM
				i := isoToCartesian(a.coord)
				// draw actors
				offset := 0.2
				if x == int(i.X+offset) && y == int(i.Y+offset) {
					if a.direction.n != 0 && a.direction.e == 0 {
						pmat = pmat.ScaledXY(pixel.ZV, pixel.V(-1, 1))
					}

					// if frame is 4 and len anims is 2
					if a.direction.n+a.direction.e < 0 {
						startingFrame = half_length
					}
					pmat = pmat.Moved(a.coord)

					a.anims[a.state][(frame%half_length+startingFrame)].Draw(batch, pmat)
				}
			}
		}
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
	pixelgl.Run(run)
}

func actorStateMachine(actors []*actor) {

	for _, a := range actors {
		for _, o := range actors {

			// _, ocharmed := o.effects["charmed"]

			// oc := 1
			// if ocharmed {
			// 	oc = -1
			// 	o.hp += 1
			// }

			// friendly is positive, enemy is negative, neutral is 0
			// if both are friendly the product of their states is positive
			// if both are hostile the product of their states is positive
			// if one is neutral the product of their states is 0
			// if they are opposed the product of their states is negative

			if a != o && a.state != dead && o.state != dead {

				d := pixel.Vec.Sub(a.coord, o.coord)
				d_square := d.X*d.X + d.Y*d.Y

				if d_square < a.prange && o.faction*a.faction < 0 {
					a.target = o
				}
			}
		}
	}

	for _, a := range actors {

		if a.state == idle {

			if a.dest.x != a.x || a.dest.y != a.y || a.target != a {
				a.state = walk
			}

		} else if a.state == walk {

			if a.dest.x == a.x && a.dest.y == a.y {
				a.state = idle
			}
			if a.target != a {
				// if in range, attack
				d := pixel.Vec.Sub(a.target.coord, a.coord)
				d_square := d.X*d.X + d.Y*d.Y

				if d_square < a.arange {
					a.state = attack
					a.direction.e = a.x - a.target.x
					a.direction.n = a.y - a.target.y
				}

				// otherwise move towards target
				path := Astar(&node{x: a.x, y: a.y}, &node{x: a.target.x, y: a.target.y}, levelData)
				if len(path) > 0 {
					if path[len(path)-1].x != a.target.x || path[len(path)-1].y != a.target.y {
						step_forward(a, path)
					}
				}

				// if no target
			} else {
				path := Astar(&node{x: a.x, y: a.y}, a.dest, levelData)
				step_forward(a, path)
			}

		} else if a.state == attack {
			a.target.hp -= 1

			if a.target.hp < 1 {
				a.state = idle
				a.target.state = dead
				a.target = a
			}
		}
	}
}

func drawHealthPlates(actors []*actor, imd *imdraw.IMDraw) {

	for _, a := range actors {
		// total length of health plate
		length := 40.0
		// number of bars to represent health (10 hp per bar)
		bars := a.maxhp / 10
		// length of a single bar
		bar_length := length / float64(bars)
		c := 20.0
		start_X := a.coord.X - c

		//percentageHealth := float64(a.hp) / float64(a.maxhp)
		if a.faction == friendly {
			imd.Color = colornames.Lightgreen
		} else if a.faction == neutral {
			imd.Color = colornames.Lightyellow
		} else if a.faction == hostile {
			imd.Color = colornames.Red
		}
		if a.hp > 0 {
			for i := 0; i < bars; i++ {

				if i*10 < a.hp && (i+1)*10 >= a.hp {
					fractionOfBar := float64((a.hp % 10)) / 10.0
					imd.Push(pixel.Vec{X: start_X + float64(i)*bar_length + 1, Y: a.coord.Y + 26.0})
					f := fractionOfBar * bar_length
					imd.Push(pixel.Vec{X: start_X + float64(i+1)*bar_length - f, Y: a.coord.Y + 26.0})
					if a.faction == friendly {
						imd.Color = colornames.Darkgreen
					} else if a.faction == neutral {
						imd.Color = colornames.Darkgoldenrod
					} else if a.faction == hostile {
						imd.Color = colornames.Darkred
					}
					imd.Push(pixel.Vec{X: start_X + float64(i+1)*bar_length - 1, Y: a.coord.Y + 26.0})
					imd.Line(3)

				} else {
					// draw the whole bar
					imd.Push(pixel.Vec{X: start_X + float64(i)*bar_length + 1, Y: a.coord.Y + 26.0})
					imd.Push(pixel.Vec{X: start_X + float64(i+1)*bar_length - 1, Y: a.coord.Y + 26.0})
					imd.Line(3)
				}
			}
		}

		// iso square
		// imd.Color = colornames.Lightpink
		// imd.Push(pixel.Vec.Sub(cartesianToIso(pixel.Vec{X: float64(player.x) - 1.5, Y: float64(player.y) + 1.5}), pixel.Vec{X: 0, Y: 7.0}))
		// imd.Push(pixel.Vec.Sub(cartesianToIso(pixel.Vec{X: float64(player.x) + 1.5, Y: float64(player.y) + 1.5}), pixel.Vec{X: 0, Y: 7.0}))
		// imd.Push(pixel.Vec.Sub(cartesianToIso(pixel.Vec{X: float64(player.x) + 1.5, Y: float64(player.y) - 1.5}), pixel.Vec{X: 0, Y: 7.0}))
		// imd.Push(pixel.Vec.Sub(cartesianToIso(pixel.Vec{X: float64(player.x) - 1.5, Y: float64(player.y) - 1.5}), pixel.Vec{X: 0, Y: 7.0}))
		// imd.Polygon(1)

	}

}
