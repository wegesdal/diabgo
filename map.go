package main

import "math/rand"

func wall_gen(x int, y int, levelData [2][32][32]uint) [2][32][32]uint {

	blocks := 6

	for blocks > 0 {
		levelData[0][x][y] = 4
		if x < 30 && x > 0 && y < 30 && y > 0 {
			r := rand.Intn(6)
			if r < 2 {
				x++
			} else if r < 4 {
				y++
			}
		}
		blocks--
	}
	return levelData
}

func generateMap(endOfTheRoad *node) [2][32][32]uint {

	var levelData = [2][32][32]uint{}

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
		levelData[0][node.x][node.y] = 1
	}
	// bake the river onto the array
	river = append(river, river_start)

	for _, node := range river {
		if levelData[0][node.x][node.y] == 1 {
			levelData[0][node.x][node.y-1] = 2
			levelData[0][node.x][node.y] = 3
			// levelData[0][node.x][node.y-1] = 2
			levelData[0][node.x][node.y+1] = 3
			levelData[0][node.x][node.y+2] = 3

		} else {
			levelData[0][node.x][node.y] = 5
			levelData[0][node.x][node.y+1] = 5
		}
	}

	// make some random trees
	for i := 0; i < rand.Intn(25); i++ {
		x := rand.Intn(31)
		y := rand.Intn(31)
		if levelData[0][x][y] == 0 {
			levelData[1][x][y] = 1
		}
	}
	return levelData
}
