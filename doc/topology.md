Topology Format
===============

`make-map` reads a topology description from a JSON file.
The format of that JSON file is:

## Top Level

    {
      "nodes": Nodes,
      "links": Links
    }
    
## Nodes

`Nodes` is either an array or and object with `Node` items.

    [ Node, ... ]
    
    // or...
    
    { NodeId: Node, ... }

## Links

`Links` is either an array or and object with `Link` items.

    [ Link, ... ]
    
    // or...
    
    { LinkId: Link, ... }

## Node

`Node` has the following format:

    {
      "id":       NodeId,
      "pos":      [int, int],
      "label":    string,
      "label_at": string,
      "class":    string,
      "style":    NodeStyle,
      "extents":  {
         "width":  int,
         "height": int
      }
    }

| Field    | Description |
| ---:     | :---        |
| id       | A unique id for the node. Required if `Nodes` is an array. |
| pos      | The position of the node in the layout grid. Required. |
| label    | The label for the node. Optional, if omitted the id is used instead. |
| label_at | The position of the label relative to the node. Values are `"c", "n", "e", "s", "w", "ne", "se", "nw", "sw"`. Optional. |
| class    | A class to assign to the node. Optional. |
| style    | Node-specific styles. Optional. |
| extents  | Specify the size of a rectangular node. The node will be centered at `pos`. Optional. |

## Link

`Link` has the following format:

    {
      "id": LinkId,
      "from": NodeId,
      "to": NodeId,
      "via": [ [int, int] ],
      "split_at": float,
      "class": string,
      "style": LinkStyle,
      "from_data": LinkData,
      "to_data": LinkData,
      "route": [ [int, int] ]
    }

| Field      | Description |
| ---:       | :---        |
| id         | A unique id for the link. Generated automatically if omitted. |
| from       | One end of the link. Required. |
| to         | The other end of the link. Required. |
| via        | A list of grid positions that the routed link must pass through. Optional. |
| split\_at  | A value between 0 and 1 describing the split point for links, 0 is the from node, 1 is the to node. Default 0.5 |
| class      | A class to assign to the link. Optional. |
| style      | Link-specific styles. Optional. |
| from\_data | Data about the link in the direction `from -> to`. Optional. |
| to\_data   | Data about the link in the direction `to -> from`. Optional. |
| route      | A list of grid positions describing a route. Not intended for use, but documented for completeness. Optional. |

Multiple links between the same two nodes are allowed.

### LinkData

`LinkData` has the following format:

    {
      "value": float,
      "label": string
    }


| Field      | Description |
| ---:       | :---        |
| value      | A value assigned to the link for the direction. Is expected to be between 0 and 1, but can be any value. Optional. |
| label      | The label for the link direction. Optional. |
