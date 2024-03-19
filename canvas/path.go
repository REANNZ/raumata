package canvas

import (
	"github.com/REANNZ/raumata/internal/f32"
	"github.com/REANNZ/raumata/vec"
)

type CommandType int

const (
	CommandClosePath CommandType = iota
	CommandMoveTo
	CommandLineTo
	CommandArcTo
)

// Path is a generic path through space.
// It can be either a line itself of the
// outline of another shape.
type Path struct {
	Element
	Data []Command
}

func NewPath() *Path {
	return &Path{}
}

// Represents a command to draw the path, which may be a
// path segment.
//
// The content of Args depends the command:
//
// * `ClosePath`: No args
// * `MoveTo`: [pos.X, pos.Y], the position to move to
// * `LineTo`: [pos.X, pos.Y], the position to draw a line to
// * `ArcTo`:  [start.X, start.Y, end.X, end.Y, radius, sweepDir]
//             start is where the arc starts, end is where the arc ends
//             radius is the radius of the circle that the arc is of,
//             sweepDir is the direction the arc is drawn in, 1 for clockwise,
//             0 for counterclockwise
type Command struct {
	Type CommandType
	Pos  vec.Vec2
	Args []float32
}

func (p *Path) addCommand(ty CommandType, pos vec.Vec2, args ...float32) {
	cmd := Command{
		Type: ty,
		Pos:  pos,
		Args: args,
	}
	p.Data = append(p.Data, cmd)
}

func (p *Path) ClosePath() *Path {
	p.addCommand(CommandClosePath, vec.Vec2{})
	return p
}

func (p *Path) MoveTo(pos vec.Vec2) *Path {
	p.addCommand(CommandMoveTo, pos, pos.X, pos.Y)
	return p
}

func (p *Path) LineTo(pos vec.Vec2) *Path {
	if len(p.Data) == 0 {
		p.addCommand(CommandMoveTo, pos, pos.X, pos.Y)
	} else {
		prev := p.Data[len(p.Data)-1]
		if !prev.Pos.ApproxEq(pos, 1e-8) {
			p.addCommand(CommandLineTo, pos, pos.X, pos.Y)
		}
	}
	return p
}

func (p *Path) Arc(start, end vec.Vec2, radius float32) *Path {
	p.LineTo(start)
	p.addCommand(CommandArcTo, end,
		start.X, start.Y, end.X, end.Y, radius, 1.0)
	return p
}

func (p *Path) ArcNeg(start, end vec.Vec2, radius float32) *Path {
	p.LineTo(start)
	p.addCommand(CommandArcTo, end,
		start.X, start.Y, end.X, end.Y, radius, 0.0)
	return p
}

// Generates a rounded corner defined by start, end and peak with the radius
func (p *Path) RoundCorner(radius float32, start, peak, end vec.Vec2) *Path {
	if radius <= 0 {
		// Special case, no arc
		return p.LineTo(start).LineTo(peak).LineTo(end)
	}

	dir1 := start.Sub(peak)
	dir2 := end.Sub(peak)

	dist1 := dir1.Length()
	dist2 := dir2.Length()

	dir1 = dir1.Div(dist1)
	dir2 = dir2.Div(dist2)

	if dir1.ApproxEq(dir2.Neg(), 1e-8) {
		// Handle the case where the three points are
		// colinear
		return p.LineTo(start).LineTo(end)
	}

	cosAngle := dir1.Dot(dir2)

	halfAngle := f32.Acos(cosAngle) / 2.0

	radiusOffset := radius / f32.Tan(halfAngle)
	offset := f32.Min(radiusOffset, dist1, dist2)

	if offset < radiusOffset {
		radius = offset * f32.Tan(halfAngle)
	}

	arcStart := start

	if offset < dist1 {
		arcStart = peak.Add(dir1.Mul(offset))
	}

	arcEnd := end
	if offset < dist2 {
		arcEnd = peak.Add(dir2.Mul(offset))
	}

	cornerDir := end.Sub(start).Norm()
	if cornerDir.Dot(dir1) > 0 {
		p.Arc(arcStart, arcEnd, radius)
	} else {
		p.ArcNeg(arcStart, arcEnd, radius)
	}

	return p.LineTo(end)
}

func (p *Path) GetAABB() *AABB {
	if p == nil {
		return nil
	}
	if len(p.Data) == 0 {
		return nil
	}

	min := p.Data[0].Pos
	max := p.Data[0].Pos

	for _, cmd := range p.Data {
		if cmd.Type == CommandClosePath {
			continue
		}
		min = min.Min(cmd.Pos)
		max = max.Max(cmd.Pos)
	}

	return NewAABB(min, max)
}

func (p *Path) Render(r Renderer) error {
	return r.RenderPath(p)
}
