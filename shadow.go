package main

// adapted from https://journal.stuffwithstuff.com/2015/09/07/what-the-hero-sees/

// type shadow struct {
// 	start int
// 	end   int
// }

// func contains(self *shadow, other *shadow) bool {
// 	return self.start <= other.start && self.end >= other.end
// }

type vec struct {
	x int
	y int
}

// var octantCoords = [8][2]*vec{
// 	{&vec{x: 0, y: -1}, &vec{x: 1, y: 0}},
// 	{&vec{x: 1, y: 0}, &vec{x: 0, y: -1}},
// 	{&vec{x: 1, y: 0}, &vec{x: 0, y: 1}},
// 	{&vec{x: 0, y: 1}, &vec{x: 1, y: 0}},
// 	{&vec{x: 0, y: 1}, &vec{x: -1, y: 0}},
// 	{&vec{x: -1, y: 0}, &vec{x: 0, y: 1}},
// 	{&vec{x: -1, y: 0}, &vec{x: 0, y: -1}},
// 	{&vec{x: 0, y: -1}, &vec{x: -1, y: 0}}}

// func add_vec(a *vec, b *vec) *vec {
// 	return &vec{x: a.x + b.x, y: a.y + b.y}
// }

// func mult_vec(v *vec, s int) *vec {
// 	return &vec{x: v.x * s, y: v.y * s}
// }

// func getProjection(row int, col int) *shadow {
// 	topLeft := col / (row + 2)
// 	bottomRight := (row + 1) / (col + 1)
// 	return &shadow{topLeft, bottomRight}
// }

// func refreshVisibility(tiles [32][32]*node, center *node) {
// 	clearVisibility(tiles)
// 	tiles[center.x][center.y].visible = true
// 	for octant := 0; octant < 8; octant++ {
// 		refreshOctant(tiles, center, octant)
// 	}

// }

// func isInShadow(shadows []*shadow, projection *shadow) bool {
// 	for i := 0; i < len(shadows); i++ {
// 		if contains(shadows[i], projection) {
// 			return true
// 		}
// 	}
// 	return false
// }

// func add_shadow(shadows []*shadow, s *shadow) []*shadow {
// 	index := 0
// 	for i := 0; i < len(shadows); i++ {
// 		index = i
// 		if shadows[i].start > s.start {
// 			break
// 		}
// 	}

// 	var overlapsPrev = ((index > 0) && (shadows[index-1].end > s.start))
// 	var overlapsNext = ((index < len(shadows)) && (shadows[index].start < s.end))

// 	if overlapsNext {
// 		if overlapsPrev {
// 			shadows[index-1].end = Max(shadows[index-1].end, shadows[index].end)
// 			shadows = removeShadow(shadows, index)
// 		} else {
// 			shadows[index].start = Min(shadows[index].start, s.start)
// 		}
// 	} else {
// 		if overlapsPrev {
// 			shadows[index-1].end = Max(shadows[index-1].end, s.end)

// 		} else {
// 			shadows = insertShadow(shadows, s, index)
// 		}
// 	}
// 	return shadows
// }

// func removeShadow(sl []*shadow, i int) []*shadow {
// 	// Remove the element at index i from a.
// 	copy(sl[i:], sl[i+1:]) // Shift a[i+1:] left one index.
// 	sl[len(sl)-1] = nil
// 	sl = sl[:len(sl)-1]
// 	return sl
// }

// func insertShadow(sl []*shadow, s *shadow, i int) []*shadow {
// 	sl = append(sl, nil)
// 	copy(sl[i+1:], sl[i:])
// 	sl[i] = s
// 	return sl
// }

// func isFullShadow(shadows []*shadow) bool {
// 	return len(shadows) == 1 && shadows[0].start == 0 && shadows[0].end == 1
// }

func clearVisibility(tiles [32][32]*node) {
	for row, _ := range tiles {
		for col, _ := range tiles[0] {
			tiles[row][col].visible = false
		}
	}
}

// func refreshOctant(tiles [32][32]*node, start *node, octant int) {

// 	rowInc := octantCoords[octant][0]
// 	colInc := octantCoords[octant][1]

// 	// fullShadow := false
// 	maxDistance := 32

// 	var shadows []*shadow

// 	for row := 1; row < maxDistance; row++ {
// 		pos := add_vec(&vec{x: start.x, y: start.y}, mult_vec(rowInc, row))
// 		// 32 is map size
// 		if !inBounds(pos, 32) {
// 			break
// 		}
// 		for col := 0; col <= row; col++ {
// 			if !inBounds(pos, 32) {
// 				break
// 			}
// 			// if fullShadow {
// 			// 	tiles[pos.x][pos.y].visible = false
// 			// } else {
// 			projection := getProjection(col, row)
// 			visible := !isInShadow(shadows, projection)
// 			tiles[pos.x][pos.y].visible = visible

// 			if !tiles[pos.x][pos.y].walkable {
// 				shadows = add_shadow(shadows, projection)
// 				// fullShadow = isFullShadow(shadows)
// 			}
// 			// }
// 			pos = add_vec(pos, colInc)
// 		}
// 	}
// }

// func inBounds(pos *vec, mapSize int) bool {
// 	return pos.x > 0 && pos.x < mapSize && pos.y > 0 && pos.y < mapSize
// }
