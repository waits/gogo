package model

import "errors"

// Point holds x and y coordinates of a point on a game board
type Point struct {
	X int
	Y int
}

// CheckLife searches for dead groups around a point and removes them
func (point Point) CheckLife(grid [][]int) ([]Point, error) {
	oppColor := 3 - grid[point.Y][point.X]
	adjacentPoints := []Point{{point.X, point.Y + 1}, {point.X + 1, point.Y}, {point.X, point.Y - 1}, {point.X - 1, point.Y}}
	capturedPieces := make([]Point, 0, 180)
	for _, p := range adjacentPoints {
		if p.X < 0 || p.X > len(grid)-1 || p.Y < 0 || p.Y > len(grid)-1 {
			continue
		} else if grid[p.Y][p.X] == oppColor {
			pieces := p.connectedDeadPieces(grid, make([]Point, 0, 180))
			if pieces != nil {
				capturedPieces = append(capturedPieces, pieces...)
				clearPoints(pieces, grid)
			}
		}
	}

	if len(capturedPieces) == 0 && point.connectedDeadPieces(grid, make([]Point, 0, 180)) != nil {
		return capturedPieces, errors.New("Illegal move: suicide")
	}

	return capturedPieces, nil
}

// Checks the points adjacent to a given point for life. Returns nil if it
// finds an empty point, otherwise it returns all connected pieces.
func (point Point) connectedDeadPieces(grid [][]int, alreadyFound []Point) []Point {
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
			alreadyFound = p.connectedDeadPieces(grid, alreadyFound)
			if alreadyFound == nil {
				return nil
			}
		default:
		}
	}

	return alreadyFound
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
