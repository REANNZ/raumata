package raumata

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/REANNZ/raumata/option"
	"github.com/REANNZ/raumata/vec"
)

type NodeId string
type LinkId string

// Represents a node on the map
type Node struct {
	Id      NodeId     `json:"id"`
	Pos     *[2]int16  `json:"pos,omitempty"`
	Label   string     `json:"label,omitempty"`
	LabelAt string     `json:"label_at,omitempty"`
	Class   string     `json:"class,omitempty"`
	Style   *NodeStyle `json:"style,omitempty"`
}

type Link struct {
	Id       LinkId       `json:"id"`
	From     NodeId       `json:"from"`
	To       NodeId       `json:"to"`
	Via      [][2]int16   `json:"via,omitempty"`
	SplitAt  *float32     `json:"split_at,omitempty"`
	Class    string       `json:"class,omitempty"`
	State    string       `json:"state,omitempty"`
	Style    *LinkStyle   `json:"style,omitempty"`
	Route    vec.Polyline `json:"route,omitempty"`
	FromData *LinkData    `json:"from_data,omitempty"`
	ToData   *LinkData    `json:"to_data,omitempty"`
}

// Data associated with a link
type LinkData struct {
	// The "value" of the link, typically link usage as a %
	Value option.Float32 `json:"value"`
	// The label for the link, typically the amount of traffic
	Label string `json:"label"`
}

// A full map topology
type Topology struct {
	Nodes map[NodeId]*Node `json:"nodes"`
	Links map[LinkId]*Link `json:"links"`
}

func (t *Topology) GetNode(id NodeId) *Node {
	return t.Nodes[id]
}

func (t *Topology) GetLink(id LinkId) *Link {
	return t.Links[id]
}

func (id NodeId) String() string {
	return string(id)
}

func (id LinkId) String() string {
	return string(id)
}

// UnmarshalJSON supports a couple different ways of representing
// a Topology
//
// It is always an object with fields: "nodes" and "links",
// However each field can either be an array of nodes/links, or an
// object with { "id": <node/link> } fields.
//
// Node ids are automatically set if the object format used, otherwise
// the must be both present and unique.
//
// Link ids, if not provided, are determined automatically from the
// "from" and "to" fields of the link.
func (t *Topology) UnmarshalJSON(data []byte) error {
	var topLevel struct {
		Nodes *json.RawMessage
		Links *json.RawMessage
	}

	err := json.Unmarshal(data, &topLevel)
	if err != nil {
		return err
	}

	nodeMap := make(map[NodeId]*Node)
	if topLevel.Nodes != nil && len(*topLevel.Nodes) > 0 {
		rawNodes := *topLevel.Nodes
		if rawNodes[0] == '[' {
			var array []*Node
			err = json.Unmarshal(rawNodes, &array)
			if err != nil {
				return err
			}

			for _, n := range array {
				if n.Id == "" {
					return errors.New("Node must have an id")
				}
				_, ok := nodeMap[n.Id]
				if ok {
					return fmt.Errorf("Duplicate node id '%s'", n.Id)
				}
				nodeMap[n.Id] = n
			}
		} else if rawNodes[0] == '{' {
			err = json.Unmarshal(rawNodes, &nodeMap)
			if err != nil {
				return err
			}
			for id, n := range nodeMap {
				n.Id = id
			}
		} else {
			return errors.New("\"nodes\" must be an array or object")
		}
		if t.Nodes == nil {
			t.Nodes = nodeMap
		} else {
			for id, node := range nodeMap {
				t.Nodes[id] = node
			}
		}
	}

	linkMap := make(map[LinkId]*Link)
	if topLevel.Links != nil && len(*topLevel.Links) > 0 {
		rawLinks := *topLevel.Links
		if rawLinks[0] == '[' {
			var array []*Link
			err = json.Unmarshal(rawLinks, &array)
			if err != nil {
				return err
			}

			for _, l := range array {
				id := l.Id
				if id == "" {
					// Automatically determine an id
					id = LinkId(fmt.Sprintf("%s-%s", l.From, l.To))

					_, ok := linkMap[id]
					n := 2
					for ok {
						id = LinkId(fmt.Sprintf("%s-%s-%d", l.From, l.To, n))
						n += 1
						_, ok = linkMap[id]
					}

					l.Id = id
				}

				_, ok := linkMap[id]
				if ok {
					return fmt.Errorf("Duplicate link id '%s'", id)
				}

				linkMap[id] = l
			}
		} else if rawLinks[0] == '{' {
			err = json.Unmarshal(rawLinks, &linkMap)
			if err != nil {
				return err
			}

			for id, link := range linkMap {
				link.Id = id
			}
		} else {
			return errors.New("\"links\" must be an array or object")
		}

		if t.Links == nil {
			t.Links = linkMap
		} else {
			for id, link := range linkMap {
				t.Links[id] = link
			}
		}
	}

	return nil
}
