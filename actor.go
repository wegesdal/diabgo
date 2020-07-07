package main

import (
	"github.com/faiface/pixel"
)

const (
	north = iota
	east
	south
	west
)

const (
	dead = iota
	idle
	attack
	walk
	cast
)

const (
	hostile = iota - 1
	neutral
	friendly
)

type actor struct {

	// actor
	x       int
	y       int
	name    string
	coord   pixel.Vec
	frame   int
	state   int
	faction int
	// movespeed float64
	effects   map[string]struct{}
	direction int
	anims     map[int][]*pixel.Sprite
}

// EFFECTS NOTES: you can get a pseudo set with the following:
// map[string] struct{}
// value, ok := yourmap[key]
// struct{}{}

func spawn_actor(x int, y int, name string, anims map[int][]*pixel.Sprite) *actor {
	var a = actor{x: x, y: y}
	a.name = name
	a.anims = anims
	a.frame = 0

	a.coord = cartesianToIso(pixel.V(float64(a.x), float64(a.y)))
	a.effects = map[string]struct{}{}
	a.state = idle

	return &a
}

func generateActorSprites(p pixel.Picture) map[int][]*pixel.Sprite {
	anim := make(map[int][]*pixel.Sprite)
	frames_per_action := 10
	directions := 4
	sprite_size := 256

	for y := 0; y < 5; y++ {
		for x := 0; x < frames_per_action*directions; x++ {
			anim[4-y] = append(anim[4-y], pixel.NewSprite(p, pixel.R(float64(sprite_size*x), float64(sprite_size*y), float64(sprite_size*(x+1)), float64(sprite_size*(y+1)))))
		}
	}
	return anim
}

func wayfind(x1 int, y1 int, x2 int, y2 int) int {
	d := 0
	if x1-x2 == 1 {
		d = 0
	} else if x1-x2 == -1 {
		d = 2
	} else if y1-y2 == 1 {
		d = 3
	} else if y1-y2 == -1 {
		d = 1
	}
	return d
}
