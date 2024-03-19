package raumata

import (
	"github.com/REANNZ/raumata/internal"
)

// Determine good placement for node labels
func PlaceLabels(topo *Topology) {
	// Records squares that are occupied
	fillGrid := internal.Grid[bool]{}

	// Record all the node positions and the positions
	// of existing labels
	for _, node := range topo.Nodes {
		if node != nil && node.Pos != nil {
			pos := internal.GridPos{
				X: node.Pos[0],
				Y: node.Pos[1],
			}
			fillGrid[pos] = true

			dir := directionFromString(node.LabelAt)

			labelAt := dir.moveGridPos(pos)

			if labelAt != pos {
				fillGrid[labelAt] = true
			}
		}
	}

	// Record all the link positions
	for _, link := range topo.Links {
		if link == nil {
			continue
		}

		for _, p := range link.Route {
			pos := internal.GridPos{
				X: int16(p.X),
				Y: int16(p.Y),
			}

			fillGrid[pos] = true
		}
	}

	// Do the label placement
	for id, node := range topo.Nodes {
		if node == nil || node.Pos == nil {
			continue
		}
		if node.LabelAt != "" {
			// Skip labels that have already been placed
			continue
		}

		pos := internal.GridPos{
			X: node.Pos[0],
			Y: node.Pos[1],
		}

		// For each valid position, calculate a score and use the position
		// with the lowest score
		bestDir := directionNone
		var bestScore float32
		for i := directionN; i <= directionNW; i++ {
			candidatePos := i.moveGridPos(pos)
			if _, ok := fillGrid[candidatePos]; !ok {
				score := evaluatePosition(candidatePos, i, id, topo.Nodes, fillGrid)
				if bestDir == directionNone || score < bestScore {
					bestScore = score
					bestDir = i
				}
			}
		}

		if bestDir != directionNone {
			node.LabelAt = bestDir.String()
			labelPos := bestDir.moveGridPos(pos)
			fillGrid[labelPos] = true
		}
	}
}

func evaluatePosition(pos internal.GridPos, dir direction, id NodeId, nodes map[NodeId]*Node, fillGrid internal.Grid[bool]) float32 {
	var score float32 = 0
	testPos := pos.ToVec()

	// Calculate the base cost for the direction
	// Favor orthogonal placement (N, E, S, or W) over
	// diagonal placement (NE, SE, SW, or NW)
	var dirCost float32
	switch dir {
	case directionN, directionE, directionS, directionW:
		dirCost = 50
	default:
		dirCost = 100
	}

	// Each node contributes to the score proportional
	// to the inverse of the distance to the node, squared
	// cost * (1/d^2)
	for nid, node := range nodes {
		if nid == id {
			continue
		}
		if node == nil || node.Pos == nil {
			continue
		}
		p := internal.GridPos{
			X: node.Pos[0],
			Y: node.Pos[1],
		}
		
		nPos := p.ToVec()
		dist := testPos.Sub(nPos).Length()
		score += dirCost / (dist*dist)
	}

	// Apply a penalty for each occupied cell around the
	// candidate position
	for d := directionN; d <= directionNW; d += 1 {
		if d == dir.Opposite() {
			// If the cell we're looking at is the node the label
			// is for, don't penalize that location
			continue
		}
		nPos := d.moveGridPos(pos)
			
		if _, ok := fillGrid[nPos]; ok {
			var penalty float32
			// If the occupied cell is to the left or right of
			// the node apply a higher penalty, since it's more
			// likely to overlap with the text.
			if d == directionE || d == directionW {
				penalty = 50
			} else {
				penalty = 5
			}
			score += penalty
		}
	}

	return score
}
