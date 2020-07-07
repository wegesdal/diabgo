package main

import (
	"math"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"golang.org/x/image/colornames"
)

type character struct {
	actor  *actor
	maxhp  int
	dest   *node
	hp     int
	prange float64
	arange float64
	target *character
}

func spawn_character(a *actor) *character {
	var c = character{actor: a}
	c.dest = &node{x: a.x, y: a.y}
	c.maxhp = 40
	c.hp = 15
	c.target = &c
	return &c
}

func step_forward(c *character, path []*node) {
	if len(path) > 0 {
		i := isoToCartesian(c.actor.coord)
		// don't update next block until close
		if math.Pow(i.X-float64(c.actor.x), 2.0)+math.Pow(i.Y-float64(c.actor.y), 2.0) < 1 {

			c.actor.direction = wayfind(c.actor.x, c.actor.y, path[len(path)-1].x, path[len(path)-1].y)
			c.actor.x = path[len(path)-1].x
			c.actor.y = path[len(path)-1].y
		}
	}
}

func characterStateMachine(characters []*character, levelData [2][32][32]uint) {

	for _, c := range characters {
		for _, o := range characters {

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

			if c != o && c.actor.state != dead && o.actor.state != dead {

				d := pixel.Vec.Sub(c.actor.coord, o.actor.coord)
				d_square := d.X*d.X + d.Y*d.Y

				if d_square < c.prange && o.actor.faction*c.actor.faction < 0 {
					c.target = o
				}
			}
		}
	}

	for _, c := range characters {

		if c.actor.state == idle {

			if c.dest.x != c.actor.x || c.dest.y != c.actor.y || c.target != c {
				c.actor.state = walk
			}

		} else if c.actor.state == walk {

			if c.dest.x == c.actor.x && c.dest.y == c.actor.y {
				c.actor.state = idle
			}
			if c.target != c {
				// if in range, attack
				d := pixel.Vec.Sub(c.target.actor.coord, c.actor.coord)
				d_square := d.X*d.X + d.Y*d.Y

				if d_square < c.arange {
					c.actor.state = attack

					c.actor.direction = wayfind(c.actor.x, c.actor.y, c.target.actor.x, c.target.actor.y)
				}

				// otherwise move towards target
				path := Astar(&node{x: c.actor.x, y: c.actor.y}, &node{x: c.target.actor.x, y: c.target.actor.y}, levelData[0])
				if len(path) > 0 {
					if path[len(path)-1].x != c.target.actor.x || path[len(path)-1].y != c.target.actor.y {
						step_forward(c, path)
					}
				}

				// if no target
			} else {
				path := Astar(&node{x: c.actor.x, y: c.actor.y}, c.dest, levelData[0])
				step_forward(c, path)
			}

		} else if c.actor.state == attack {
			c.target.hp -= 1

			if c.target.hp < 1 {
				c.actor.state = idle
				c.target.actor.frame = 0
				c.target.actor.state = dead
				c.target = c
			}
		}
	}
}

func drawHealthPlates(characters []*character, imd *imdraw.IMDraw) {

	for _, c := range characters {
		// total length of health plate
		length := 40.0
		// number of bars to represent health (10 hp per bar)
		bars := c.maxhp / 10
		// length of a single bar
		bar_length := length / float64(bars)
		z := 20.0
		start_X := c.actor.coord.X - z

		//percentageHealth := float64(a.hp) / float64(a.maxhp)
		if c.actor.faction == friendly {
			imd.Color = colornames.Lightgreen
		} else if c.actor.faction == neutral {
			imd.Color = colornames.Lightyellow
		} else if c.actor.faction == hostile {
			imd.Color = colornames.Red
		}
		if c.hp > 0 {
			for i := 0; i < bars; i++ {
				verticalOffset := 192.0

				if i*10 < c.hp && (i+1)*10 >= c.hp {

					fractionOfBar := float64((c.hp % 10)) / 10.0
					imd.Push(pixel.Vec{X: start_X + float64(i)*bar_length + 1, Y: c.actor.coord.Y + verticalOffset})
					f := fractionOfBar * bar_length
					imd.Push(pixel.Vec{X: start_X + float64(i+1)*bar_length - f, Y: c.actor.coord.Y + verticalOffset})
					if c.actor.faction == friendly {
						imd.Color = colornames.Darkgreen
					} else if c.actor.faction == neutral {
						imd.Color = colornames.Darkgoldenrod
					} else if c.actor.faction == hostile {
						imd.Color = colornames.Darkred
					}
					imd.Push(pixel.Vec{X: start_X + float64(i+1)*bar_length - 1, Y: c.actor.coord.Y + verticalOffset})
					imd.Line(3)

				} else {
					// draw the whole bar
					imd.Push(pixel.Vec{X: start_X + float64(i)*bar_length + 1, Y: c.actor.coord.Y + verticalOffset})
					imd.Push(pixel.Vec{X: start_X + float64(i+1)*bar_length - 1, Y: c.actor.coord.Y + verticalOffset})
					imd.Line(3)
				}
			}

			// iso square
			// imd.Color = colornames.Lightpink
			// imd.Push(pixel.Vec.Add(a.coord, cartesianToIso(pixel.Vec{X: -1.5, Y: -1.5})))
			// imd.Push(pixel.Vec.Add(a.coord, cartesianToIso(pixel.Vec{X: -1.5, Y: 1.5})))
			// imd.Push(pixel.Vec.Add(a.coord, cartesianToIso(pixel.Vec{X: 1.5, Y: 1.5})))
			// imd.Push(pixel.Vec.Add(a.coord, cartesianToIso(pixel.Vec{X: 1.5, Y: -1.5})))
			// imd.Polygon(1)
		}

	}

}