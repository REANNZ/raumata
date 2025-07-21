/*
MakeMap generates a map from a topology.

Usage:

	make-map [flags] [input [output]]

The flags are:

		-c path
		    Read config from the JSON-formatted file at path.
		-dumpconf
		    Dump the config as JSON to stdout and exit.
		-h, -help
		    Print out full help
		-no-spread-links
		    Don't spread links out when routing

If the input arg is not set, then the topology is read from standard input.
If the output arg is not set, then the output is written to standard output.
*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/REANNZ/raumata"
	"github.com/REANNZ/raumata/canvas"
	"github.com/REANNZ/raumata/vec"
)

var (
	configPath    string = ""
	help          bool   = false
	dumpConf      bool   = false
	noSpreadLinks bool   = false
)

func init() {
	flag.StringVar(&configPath, "c", "", "path to a config file in JSON format")
	flag.BoolVar(&help, "h", false, "")
	flag.BoolVar(&help, "help", false, "")
	flag.BoolVar(&dumpConf, "dumpconf", false, "")
	flag.BoolVar(&noSpreadLinks, "no-spread-links", false, "")
}

func main() {
	flag.Parse()

	if help {
		printHelp()
		return
	}

	os.Exit(run())
}

func run() int {

	renderConfig := raumata.DefaultRenderConfig()
	if configPath != "" {
		f, err := os.Open(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening config file %s: %s\n",
				configPath, err)
			return 1
		}

		decoder := json.NewDecoder(f)
		err = decoder.Decode(renderConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing config: %s\n", err)
			return 1
		}
	}

	if dumpConf {
		dumpConfig(renderConfig)
		return 0
	}

	var in io.Reader = os.Stdin

	if flag.NArg() > 0 {
		input := flag.Arg(0)
		if input != "-" {
			f, err := os.Open(input)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error opening file %s: %s\n",
					input, err)
				return 1
			}
			in = f
		}
	}

	var out io.Writer = os.Stdout
	var tmpFile *os.File = nil
	var dstFilename string

	defer func() {
		if tmpFile != nil {
			os.Remove(tmpFile.Name())
		}
	}()

	if flag.NArg() > 1 {
		name := flag.Arg(1)
		if name != "-" {
			dstFilename = name
			f, err := os.CreateTemp("", "map.*")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error opening temporary file %s: %s\n",
					f.Name(), err)
				return 1
			}
			out = f
			tmpFile = f
		}
	}

	topo := raumata.Topology{}

	decoder := json.NewDecoder(in)
	if err := decoder.Decode(&topo); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing topology: %s\n", err)
		return 1
	}

	linkRouter := raumata.NewLinkRouter(&topo)
	linkRouter.SpreadLinks = !noSpreadLinks
	min, max := linkRouter.GetExtents()
	linkRouter.SetExtents(int(min.X-1), int(min.Y-1), int(max.X+1), int(max.Y+1))
	linkRouter.RouteLinks()

	raumata.PlaceLabels(&topo)

	renderer := raumata.NewRendererWithConfig(renderConfig)
	c := canvas.NewCanvas()
	c.Margin = vec.Vec2{X: 10, Y: 10}

	err := renderer.RenderTopologyToCanvas(&topo, c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering topology: %s\n", err)
		return 1
	}

	svgRenderer := canvas.NewSVGRenderer(out)
	svgRenderer.Indent = 2

	if err := c.Render(svgRenderer); err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering to SVG: %s\n", err)
		return 1
	}

	if tmpFile != nil {
		if err := os.Rename(tmpFile.Name(), dstFilename); err != nil {
			fmt.Fprintf(os.Stderr, "Error moving output to final location: %s\n", err)
			return 1
		}
		tmpFile = nil
	}

	return 0
}

func printHelp() {

	usage := `MakeMap generates a map from a topology.

Usage:

    make-map [flags] [input [output]]

The flags are:

    -c path
          Read config from the JSON-formatted file at path.
    -dumpconf
          Dump the config as JSON to stdout and exit.
    -h, -help
        Print out full help
    -no-spread-links
        Don't spread links out when routing

If input isn't set, or has the value '-', the topology is read
from standard input.
If output isn't set, or has the value '-' the map is written
to standard output.

Otherwise, the arguments are paths to to the input and output files.
`

	io.WriteString(os.Stderr, usage)
}

func dumpConfig(conf *raumata.RenderConfig) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	encoder.Encode(conf)
}
