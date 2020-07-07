package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
)

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

func handleTerminalInput(player *character, txt *text.Text, input string) string {
	if player.actor.state == activate {
		txt.WriteString(win.Typed())
		input += win.Typed()
		if win.JustPressed(pixelgl.KeyEnter) || win.Repeated(pixelgl.KeyEnter) {
			switch input {
			case "foo":
				txt.WriteRune('\n')
				txt.WriteString("bar")
			case "heal":
				player.hp = player.maxhp
				txt.WriteRune('\n')
				txt.WriteString("completed")
			default:
				txt.WriteRune('\n')
				txt.WriteString(input + " is not defined")
			}
			input = ""
			txt.WriteRune('\n')
			txt.WriteString("> ")

		}
	}
	return input
}

func renderTerminalText(player *character, txt *text.Text, input string) string {

	if player.actor.state == activate {
		txt.Draw(win, pixel.IM.Moved(pixel.Vec{X: player.actor.coord.X + 22.0, Y: player.actor.coord.Y + 90.0}.Sub(txt.Bounds().Min)))
	} else {
		input = ""
		txt.Clear()
		txt.Color = pixel.RGBA{R: 1.0, G: 0.6, B: 0, A: 1.0}
		txt.WriteString("> ")
	}
	return input
}
