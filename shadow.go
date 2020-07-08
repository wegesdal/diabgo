package main

// adapted from https://journal.stuffwithstuff.com/2015/09/07/what-the-hero-sees/

type shadow struct {
	start int
	end   int
}

func contains(self *shadow, other *shadow) bool {
	return self.start <= other.start && self.end >= other.end
}

type shadowLine struct {
	shadows []*shadow
}

func projectTile(row int, col int) *shadow {
	topLeft := col / (row + 2)
	bottomRight := (col + 1) / (row + 1)
	return &shadow{topLeft, bottomRight}
}

func refreshVisibility(tiles [32][32]*node, center *node) {
	clearVisibility(tiles)
	for octant := 0; octant < 8; octant++ {
		refreshOctant(tiles, center, octant)
	}
}

func transformOctant(row int, col int, octant int) *node {
	var (
		x int
		y int
	)
	switch octant {
	case 0:
		x, y = col, -row
	case 1:
		x, y = row, -col
	case 2:
		x, y = row, col
	case 3:
		x, y = col, row
	case 4:
		x, y = -col, row
	case 5:
		x, y = -row, col
	case 6:
		x, y = -row, -col
	case 7:
		x, y = -col, -row
	}
	return &node{x: x, y: y}
}

func isInShadow(sl *shadowLine, projection *shadow) bool {
	for i := 0; i < len(sl.shadows); i++ {
		if contains(sl.shadows[i], projection) {
			return true
		}
	}
	return false
}

func add_shadow(sl *shadowLine, s *shadow) []*shadow {
	index := 0
	for i := 0; i < len(sl.shadows); i++ {
		index = i
		if sl.shadows[i].start >= s.start {
			break
		}
	}

	var overlappingPrevious *shadow

	if index > 0 && sl.shadows[index-1].end > s.start {
		overlappingPrevious = sl.shadows[index-1]
	}

	var overlappingNext *shadow

	if index < len(sl.shadows) && sl.shadows[index].start < s.end {
		overlappingNext = sl.shadows[index]
	}
	if overlappingNext != nil {
		if overlappingPrevious != nil {
			overlappingPrevious.end = overlappingNext.end
			sl.shadows = removeShadow(sl.shadows, index)
		} else {
			overlappingNext.start = s.start
		}
	} else {
		if overlappingPrevious != nil {
			overlappingPrevious.end = s.end
		} else {
			// write insert code here
			sl.shadows = insertShadow(sl.shadows, s, index)
		}
	}
	return sl.shadows
}

func removeShadow(sl []*shadow, i int) []*shadow {
	// Remove the element at index i from a.
	copy(sl[i:], sl[i+1:]) // Shift a[i+1:] left one index.
	sl[len(sl)-1] = nil
	sl = sl[:len(sl)-1]
	return sl
}

func insertShadow(sl []*shadow, s *shadow, i int) []*shadow {
	sl = append(sl, nil)
	copy(sl[i+1:], sl[i:])
	sl[i] = s
	return sl
}

func isFullShadow(sl *shadowLine) bool {
	return len(sl.shadows) == 1 && sl.shadows[0].start == 0 && sl.shadows[0].end == 1
}

func clearVisibility(tiles [32][32]*node) {
	for row, _ := range tiles {
		for col, _ := range tiles[0] {
			tiles[row][col].visible = false
		}
	}
}

func refreshOctant(tiles [32][32]*node, start *node, octant int) {
	tiles[start.x][start.y].visible = true
	line := shadowLine{}
	fullShadow := false
	maxDistance := 10
	for row := 1; row < maxDistance; row++ {
		t := transformOctant(row, 0, octant)
		x, y := start.x+t.x, start.y+t.y
		// 32 is map size
		if !inBounds(x, y, 32) {
			break
		}
		for col := 0; col <= row; col++ {
			t = transformOctant(row, col, octant)
			x, y = start.x+t.x, start.y+t.y
			if !inBounds(x, y, 32) {
				break
			}
			if fullShadow {
				tiles[x][y].visible = false
			} else {
				projection := projectTile(row, col)
				visible := !isInShadow(&line, projection)
				tiles[x][y].visible = visible

				if visible && !tiles[x][y].walkable {
					line.shadows = add_shadow(&line, projection)
					fullShadow = isFullShadow(&line)
				}
			}
		}
	}
}

func inBounds(x int, y int, mapSize int) bool {
	return x > 0 && x < mapSize && y > 0 && y < mapSize
}
