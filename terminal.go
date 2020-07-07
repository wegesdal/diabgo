package main

import "github.com/faiface/pixel"

func terminalStateMachine(actors []*actor) {

	var player *actor
	for _, p := range actors {
		if p.name == "player" {
			player = p
		}
	}

	activation_radius := 2000.0

	for _, a := range actors {
		if a.name == "terminal" {

			if player != nil {
				d := pixel.Vec.Sub(player.coord, a.coord)
				d_square := d.X*d.X + d.Y*d.Y

				if d_square < activation_radius {
					if player.state == idle {
						a.state = activate
						player.state = activate
						player.x = a.x - 1
						player.y = a.y
						player.direction = south
						a.frame = 0
					} else {
						if a.frame < 9 {
							a.frame++
						}
					}
				} else {
					if a.frame > 0 {
						a.frame--
					}
				}
			}
		}
	}
}
