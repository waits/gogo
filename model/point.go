package model

type Point struct {
	X int
	Y int
}

// Checks the points adjacent to a given point. On an empty space it returns
// true, on an opposing color it continues, and on a like color it recurses.
func deadPiecesAdjacentTo(point Point, grid [][]int8, alreadyFound []Point) []Point {
	color := grid[point.Y][point.X]
	adjacentPoints := []Point{{point.X, point.Y + 1}, {point.X + 1, point.Y}, {point.X, point.Y - 1}, {point.X - 1, point.Y}}
	alreadyFound = append(alreadyFound, point)

	for _, p := range adjacentPoints {
		if pointInSet(p, alreadyFound) {
			continue
		}

		if p.X < 0 || p.X > len(grid)-1 || p.Y < 0 || p.Y > len(grid)-1 {
			continue
		}

		if grid[p.Y][p.X] == 0 {
			return nil
		}

		if grid[p.Y][p.X] == color {
			alreadyFound = deadPiecesAdjacentTo(p, grid, alreadyFound)
		}
	}

	return alreadyFound
}

func piecesAdjacentTo(point Point, color int8, grid [][]int8) []Point {
	adjacentPoints := []Point{{point.X, point.Y + 1}, {point.X + 1, point.Y}, {point.X, point.Y - 1}, {point.X - 1, point.Y}}
	result := make([]Point, 0, 4)
	for _, p := range adjacentPoints {
		if p.X < 0 || p.X > len(grid)-1 || p.Y < 0 || p.Y > len(grid)-1 {
			continue
		} else if grid[p.Y][p.X] == color {
			result = append(result, Point{p.X, p.Y})
		}
	}
	return result
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
