package main

import (
	"math/rand"

	"github.com/faiface/pixel"
)

func wall_gen(x int, y int, levelData [2][32][32]*node) [2][32][32]*node {

	blocks := 6

	for blocks > 0 {
		levelData[0][x][y].tile = 4
		levelData[0][x][y].walkable = false

		if x < 30 && x > 0 && y < 30 && y > 0 {
			d6 := rand.Intn(6)
			if d6 < 2 {
				x++
			} else if d6 < 4 {
				y++
			}
		}
		blocks--
	}
	return levelData
}

// TODO: Clean up art file and import structure
func generateTiles(pic pixel.Picture) [2][]*pixel.Sprite {
	var tiles [2][]*pixel.Sprite
	// ground
	tiles[0] = append(tiles[0], pixel.NewSprite(pic, pixel.R(0, 64, tileSize, 128)))

	// road
	tiles[0] = append(tiles[0], pixel.NewSprite(pic, pixel.R(0, 128, tileSize, 192)))

	// bridge
	tiles[0] = append(tiles[0], pixel.NewSprite(pic, pixel.R(64, 256, 128, 192)))

	tiles[0] = append(tiles[0], pixel.NewSprite(pic, pixel.R(0, 256, tileSize, 192)))

	// block
	tiles[0] = append(tiles[0], pixel.NewSprite(pic, pixel.R(0, 0, tileSize, 64)))

	// water
	tiles[0] = append(tiles[0], pixel.NewSprite(pic, pixel.R(64, 128, 128, 192)))

	// the 0th doodad is nil to allow direct grid assignment (empty array is 0)
	tiles[1] = append(tiles[1], nil)
	tiles[1] = append(tiles[1], pixel.NewSprite(pic, pixel.R(128, 64, tileSize+128, 192)))
	return tiles
}

func generateMap() ([2][32][32]*node, *node) {
	endOfTheRoad := &node{x: rand.Intn(31), y: 31}
	var levelData = [2][32][32]*node{}

	for z := 0; z < 2; z++ {
		for x := 0; x < 32; x++ {
			for y := 0; y < 32; y++ {
				levelData[z][x][y] = &node{x: x, y: y, tile: 0}
				if levelData[0][x][y].tile < 4 {
					levelData[0][x][y].walkable = true
				} else {
					levelData[0][x][y].walkable = false
				}
			}
		}
	}

	// make some walls
	for i := 0; i < 10; i++ {
		start_x := rand.Intn(25) + 4
		start_y := rand.Intn(25) + 4
		levelData = wall_gen(start_x, start_y, levelData)
	}

	// generate a path
	road_start := &node{x: rand.Intn(31), y: 0}

	road := Astar(road_start, endOfTheRoad, levelData[0])

	// generate a river
	river_start := &node{x: 0, y: rand.Intn(31)}
	river := Astar(river_start, &node{x: 31, y: rand.Intn(31)}, levelData[0])

	// bake the road onto the array
	for _, node := range road {
		levelData[0][node.x][node.y].tile = 1
		levelData[0][node.x+1][node.y].tile = 1
		levelData[0][node.x][node.y].walkable = true
		levelData[0][node.x+1][node.y].walkable = true
	}
	// bake the river onto the array
	river = append(river, river_start)

	for _, node := range river {
		if levelData[0][node.x][node.y].tile == 1 {
			levelData[0][node.x][node.y].tile = 3
			levelData[0][node.x][node.y+1].tile = 3
			levelData[0][node.x][node.y+2].tile = 3
		} else {
			levelData[0][node.x][node.y].tile = 5
			levelData[0][node.x][node.y].walkable = false
			levelData[0][node.x][node.y+1].tile = 5
			levelData[0][node.x][node.y+1].walkable = false
		}
	}

	// make some random trees
	for i := 0; i < rand.Intn(25); i++ {
		x := rand.Intn(31)
		y := rand.Intn(31)
		if levelData[0][x][y].tile == 0 {
			levelData[1][x][y].tile = 1
		}
	}

	return levelData, endOfTheRoad
}
