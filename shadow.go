package main

// type shadow struct {
// 	start int
// 	end   int
// }

// func contains(self *shadow, other *shadow) bool {
// 	return self.start <= other.start && self.end >= other.end
// }

// type shadowLine struct {
// 	shadows []*shadow
// }

// func projectTile(row int, col int) *shadow {
// 	topLeft := col - 1/(row+2)
// 	bottomRight := (col + 1) / (row + 1)
// 	return &shadow{topLeft, bottomRight}
// }

// func refreshVisibility(center *node) {
// 	for octant := 0; octant < 7; octant++ {
// 		refreshOctant(center, octant)
// 	}
// }

// func transformOctant(row int, col int, octant int) *node {
// 	var (
// 		x int
// 		y int
// 	)
// 	switch octant {
// 	case 0:
// 		x, y = col, -row
// 	case 1:
// 		x, y = row, -col
// 	case 2:
// 		x, y = row, col
// 	case 3:
// 		x, y = col, row
// 	case 4:
// 		x, y = -col, row
// 	case 5:
// 		x, y = -row, col
// 	case 6:
// 		x, y = -row, -col
// 	case 7:
// 		x, y = -col, -row
// 	}
// 	return &node{x: x, y: y}
// }

// func isInShadow(sl, projection) {
// 	var r bool
// 	for i := 0; i< len(sl.shadows); i++ {
// 		if contains(sl.shadows[i], projection) {
// 			r = true
// 		}
// 	}
// 	r = false
// 	return r
// }

// func add_shadow(sl *shadowLine, shadow *shadow) {
// index := 0
// for i, s := range sl.shadows {
// 	if s.start >= shadow.start {
// 		index = i
// 		break
// 	}
// }

// var overlappingPrevious shadow

// if index > 0 && sl.shadows[index -1].finish > shadow.start {
// 	overlappingPrevious = shadows[index-1]
// }

// var overlappingNext *shadow

// if index < len(sl.shadows - 1) && sl.shadows[index].start < shadow.finish {
// 	overlappingNext = shadows[index]
// }
// if overlappingNext != nil {
// 	if overlappingPrevious != nil {
// 	overlappingPrevious.finish = overlappingNext.finish
// 	// write remove shadow function here
// 	removeShadow(sl.shadows, index)
// 	} else {
// 		overlappingNext.start = shadow.start
// 	}
// } else {
// 	if overlappingPrevious != nil {
// 		overlappingPrevious.finish = shadow.finish
// 	} else {
// 		// write insert code here
// 		insert(sl.shadows, index, shadow)
// 	}
// }
// }

// func isFullShadow(sl *shadowLine) bool {
// 	return sl.shadows == 1 && sl.shadows[1].start == 0 && shadows[1].finish == 1
// }

// func refreshOctant(start *node, octant int) {

// 	tiles[start.y][start.x].isVisible = true

// 	// create new blank array the size of map and set everything to true

//     line = shadowLine{}
// 	fullShadow := false

// 	for row := 0; row < 5; row++ {
// 		t := transformOctant(row, 0, octant)
// 		pos = Vec(start.x + t.x, start.y + t.y)
// 		// 32 is map size
//         inBounds := pos.x > 0 && pos.x < 32 && pos.y > 0 && pos.y < 32
//         if inBounds {
//             for col := 0, col < row; col++ {
//                 t = transformOctant(row, col, octant)
//                 pos = Vec(start.x + t.x, start.y + t.y)
//                 if (pos.x > 0 && pos.x < 32 && pos.y > 0 && pos.y < 32) {
//                 if fullShadow {
//                     tiles[pos.y][pos.x].isVisible = false
//                 } else {
//                     projection = projectTile(row, col)
//                     visible = !isInShadow(line, projection)
//                     tiles[pos.y][pos.x].isVisible = visible
//                     if visible && tiles[pos.y][pos.x].isWall{
//                         line:add(projection)
//                         fullShadow = isFullShadow(line)
// 							}
// 						}
// 					}
// 				}
// 			}
// 		}
// 		return line.shadows
// 	}
