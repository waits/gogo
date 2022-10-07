package model

import "errors"

const (
	empty = iota
	black = iota
	white = iota
	mixed = iota
)

// Point holds x and y coordinates of a point on a game board
type Point struct {
	X int
	Y int
}

// CheckLife searches for dead groups around a point and removes them
func (point Point) CheckLife(grid [][]int) ([]Point, error) {
	oppColor := 3 - grid[point.Y][point.X]
	adjacentPoints := []Point{{point.X, point.Y + 1}, {point.X + 1, point.Y}, {point.X, point.Y - 1}, {point.X - 1, point.Y}}
	capturedStones := make([]Point, 0, 180)
	for _, p := range adjacentPoints {
		if p.X < 0 || p.X > len(grid)-1 || p.Y < 0 || p.Y > len(grid)-1 {
			continue
		} else if grid[p.Y][p.X] == oppColor {
			stones := p.connectedDeadStones(grid, make([]Point, 0, 180))
			if stones != nil {
				capturedStones = append(capturedStones, stones...)
				clearPoints(stones, grid)
			}
		}
	}

	if len(capturedStones) == 0 && point.connectedDeadStones(grid, make([]Point, 0, 180)) != nil {
		return capturedStones, errors.New("Illegal move: suicide")
	}

	return capturedStones, nil
}

// Checks the points adjacent to a given point for life. Returns nil if it
// finds an empty point, otherwise it returns all connected stones.
func (point Point) connectedDeadStones(grid [][]int, alreadyFound []Point) []Point {
	color := grid[point.Y][point.X]
	adjacentPoints := []Point{{point.X, point.Y - 1}, {point.X + 1, point.Y}, {point.X, point.Y + 1}, {point.X - 1, point.Y}}
	alreadyFound = append(alreadyFound, point)

	for _, p := range adjacentPoints {
		if p.inSet(alreadyFound) {
			continue
		} else if p.X < 0 || p.X > len(grid)-1 || p.Y < 0 || p.Y > len(grid)-1 {
			continue
		}

		switch grid[p.Y][p.X] {
		case 0:
			return nil
		case color:
			alreadyFound = p.connectedDeadStones(grid, alreadyFound)
			if alreadyFound == nil {
				return nil
			}
		default:
		}
	}

	return alreadyFound
}

// Searches for adjacent empty spaces. Returns a slice of empty points and the owning color.
func (point Point) searchArea(grid [][]int, found []Point, color int) ([]Point, int) {
	adjacentPoints := []Point{{point.X, point.Y - 1}, {point.X + 1, point.Y}, {point.X, point.Y + 1}, {point.X - 1, point.Y}}
	found = append(found, point)

	for _, p := range adjacentPoints {
		if p.inSet(found) {
			continue
		} else if p.X < 0 || p.X > len(grid)-1 || p.Y < 0 || p.Y > len(grid)-1 {
			continue
		}

		switch grid[p.Y][p.X] {
		case empty:
			found, color = p.searchArea(grid, found, color)
		case color:
			continue
		default:
			if color == empty {
				color = grid[p.Y][p.X]
			} else {
				color = mixed
			}
		}
	}

	return found, color
}

// Returns true if a matching point is found in the slice
func (point Point) inSet(set []Point) bool {
	for _, member := range set {
		if member == point {
			return true
		}
	}

	return false
}

// Resets each point in the slice to 0
func clearPoints(points []Point, grid [][]int) {
	for _, p := range points {
		grid[p.Y][p.X] = 0
	}
}
