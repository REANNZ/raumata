package raumata_test

import (
	"encoding/json"
	"testing"

	. "github.com/REANNZ/raumata"
)

func TestUnmarshalTopology(t *testing.T) {
	jsonBlob := `{
  "nodes": {
    "a": {
      "pos": [0, 0]
    },
    "b": {
      "pos": [1, 1]
    },
    "c": {
      "pos": [3, 1]
    }
  },
  "links": {
    "a-b": {
      "from": "a",
      "to": "b"
    },
    "b-c": {
      "from": "b",
      "to": "c"
    },
    "c-a": {
      "from": "c",
      "to": "a"
    }
  }
}`

	topo := Topology{}

	err := json.Unmarshal([]byte(jsonBlob), &topo)
	if err != nil {
		t.Errorf("Error unmarshalling into Topology: %s", err)
		return
	}
}
