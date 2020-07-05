package main

import (
	"math"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"golang.org/x/image/colornames"
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
