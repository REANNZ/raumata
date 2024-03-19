package raumata

import (
	"strings"

	"github.com/REANNZ/raumata/internal"
	"github.com/REANNZ/raumata/vec"
)

// Helper type for handling compass directions
type direction int

const (
	directionNone direction = iota
	directionN
	directionNE
	directionE
	directionSE
	directionS
	directionSW
	directionW
	directionNW
)

func directionFromString(s string) direction {
	switch strings.ToLower(s) {
	case "n", "north":
		return directionN
	case "ne", "northeast", "north-east":
		return directionNE
	case "e", "east":
		return directionE
	case "se", "southeast", "south-east":
		return directionSE
	case "s", "south":
		return directionS
	case "sw", "southwest", "south-west":
		return directionSW
	case "w", "west":
		return directionW
	case "nw", "northwest", "north-west":
		return directionNW
	default:
		return directionNone
	}
}

func (d direction) Opposite() direction {
	switch d {
	case directionN:
		return directionS
	case directionNE:
		return directionSW
	case directionE:
		return directionW
	case directionSE:
		return directionNW
	case directionS:
		return directionN
	case directionSW:
		return directionNE
	case directionW:
		return directionE
	case directionNW:
		return directionSE
	default:
		return d
	}
}

// Returns the direction as a vector.
// Y-values increase as you go south
func (d direction) AsVec() vec.Vec2 {
	switch d {
	case directionN:
		return vec.Vec2{X: 0, Y: -1}
	case directionNE:
		return vec.Vec2{X: 1, Y: -1}
	case directionE:
		return vec.Vec2{X: 1, Y: 0}
	case directionSE:
		return vec.Vec2{X: 1, Y: 1}
	case directionS:
		return vec.Vec2{X: 0, Y: 1}
	case directionSW:
		return vec.Vec2{X: -1, Y: 1}
	case directionW:
		return vec.Vec2{X: -1, Y: 0}
	case directionNW:
		return vec.Vec2{X: -1, Y: -1}
	default:
		return vec.Vec2{}
	}
}

// Returns the given grid position moved by the direction
func (d direction) moveGridPos(p internal.GridPos) internal.GridPos {
	v := d.AsVec()

	return internal.GridPos{
		X: p.X + int16(v.X),
		Y: p.Y + int16(v.Y),
	}
}

func (d direction) String() string {
	switch d {
	case directionN:
		return "n"
	case directionNE:
		return "ne"
	case directionE:
		return "e"
	case directionSE:
		return "se"
	case directionS:
		return "s"
	case directionSW:
		return "sw"
	case directionW:
		return "w"
	case directionNW:
		return "nw"
	default:
		return ""
	}
}
