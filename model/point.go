package model

import "errors"

// import "fmt"

type Point struct {
	X int
	Y int
}

// Checks the points adjacent to a given point for life. Returns nil if it
// finds an empty point, otherwise it returns all connected pieces.
func deadPiecesConnectedTo(point Point, grid [][]int8, alreadyFound []Point) []Point {
	color := grid[point.Y][point.X]
	adjacentPoints := []Point{{point.X, point.Y - 1}, {point.X + 1, point.Y}, {point.X, point.Y + 1}, {point.X - 1, point.Y}}
	alreadyFound = append(alreadyFound, point)

// 	fmt.Printf("Searching around %v\n", point)
	for _, p := range adjacentPoints {
		if pointInSet(p, alreadyFound) {
// 			fmt.Printf("%v [already in set]\n", p)
			continue
		} else if p.X < 0 || p.X > len(grid)-1 || p.Y < 0 || p.Y > len(grid)-1 {
// 			fmt.Printf("%v [out of bounds]\n", p)
			continue
		}

		switch grid[p.Y][p.X] {
			case 0:
// 				fmt.Printf("%v [empty]\n", p)
				return nil
			case color:
// 				fmt.Printf("%v [ally]\n", p)
				alreadyFound = deadPiecesConnectedTo(p, grid, alreadyFound)
				if alreadyFound == nil {
					return nil
				}
			default:
// 				fmt.Printf("%v [opponent]\n", p)
		}
	}

	return alreadyFound
}

// Searches for dead groups around a point and removes them
func removeDeadPiecesAround(point Point, grid [][]int8) (int, error) {
	oppColor := 3 - grid[point.Y][point.X]
	adjacentPoints := []Point{{point.X, point.Y + 1}, {point.X + 1, point.Y}, {point.X, point.Y - 1}, {point.X - 1, point.Y}}
	captured := 0
	for _, p := range adjacentPoints {
		if p.X < 0 || p.X > len(grid)-1 || p.Y < 0 || p.Y > len(grid)-1 {
			continue
		} else if grid[p.Y][p.X] == oppColor {
			pieces := deadPiecesConnectedTo(p, grid, make([]Point, 0, 180))
			if pieces != nil {
				captured += len(pieces)
				clearPoints(pieces, grid)
			}
		}
	}

	if captured == 0 && deadPiecesConnectedTo(point, grid, make([]Point, 0, 180)) != nil {
		return 0, errors.New("Illegal move: suicide")
	}

	return captured, nil
}

// Returns true if a matching point is found in the slice
func pointInSet(p Point, set []Point) bool {
	for _, member := range set {
		if member == p {
			return true
		}
	}

	return false
}

// Resets each point in the slice to 0
func clearPoints(points []Point, grid [][]int8) {
	for _, p := range points {
		grid[p.Y][p.X] = 0
	}
}
