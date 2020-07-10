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
	activate = iota
	dead
	cast
	walk
	attack
	idle
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
	a.direction = 0

	a.coord = cartesianToIso(pixel.V(float64(a.x), float64(a.y)))
	a.effects = map[string]struct{}{}
	a.state = idle
	return &a
}

func generateActorSprites(p pixel.Picture, num_rows int, size int) map[int][]*pixel.Sprite {
	anim := make(map[int][]*pixel.Sprite)
	frames_per_action := 10
	directions := 4
	for y := 0; y < num_rows; y++ {
		for x := 0; x < frames_per_action*directions; x++ {
			anim[4-y] = append(anim[4-y], pixel.NewSprite(p, pixel.R(float64(size*x), float64(size*y), float64(size*(x+1)), float64(size*(y+1)))))
		}
	}
	return anim
}

func generateCharacterSprites(p pixel.Picture, size int) map[int][]*pixel.Sprite {
	anim := make(map[int][]*pixel.Sprite)
	num_poses := 6
	num_angles := 8
	num_frames := 10
	for i := 0; i < num_poses; i++ {
		for a := 0; a < num_angles; a++ {
			for f := 0; f < num_frames; f++ {
				y_offset := i*size*4 + (a/2)*size
				x_offset := (a%2)*num_frames*size + f*size
				anim[i] = append(anim[i], pixel.NewSprite(p, pixel.R(float64(x_offset), float64(y_offset), float64(x_offset+size), float64(y_offset+size))))
			}
		}
	}
	return anim
}

func wayfind(x1 int, y1 int, x2 int, y2 int) int {
	d := 0
	xy_diff := vec{x: x1 - x2, y: y1 - y2}

	switch {
	case xy_diff.x == 1 && xy_diff.y == 0:
		d = 4
	case xy_diff.x == -1 && xy_diff.y == 0:
		d = 0
	case xy_diff.x == 0 && xy_diff.y == 1:
		d = 6
	case xy_diff.x == 0 && xy_diff.y == -1:
		d = 2
	case xy_diff.x == 1 && xy_diff.y == 1:
		d = 7
	case xy_diff.x == -1 && xy_diff.y == 1:
		d = 1
	case xy_diff.x == 1 && xy_diff.y == -1:
		d = 5
	case xy_diff.x == -1 && xy_diff.y == -1:
		d = 3
	}
	return d
}
