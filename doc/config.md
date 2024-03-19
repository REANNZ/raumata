Config Format
===============

`make-map` optionally reads configuration from a JSON file.
The format of that JSON file is:

## Top Level

    {
      "min-node-sep": float,
      "node-style": NodeStyle,
      "node-styles": {
        string: NodeStyle, ...
      },
      "link-style": LinkStyle,
      "link-styles": {
        string: LinkStyle, ...
      },
      "node-label-style": NodeLabelStyle,
      "link-label-style": LinkLabelStyle,
      "link-color-scale": ColorScale
    }

| Field            | Description |
| ---:             | :---        |
| min-node-sep     | The minimum distance between adjacent nodes of the same size. Higher values spread nodes out more. |
| node-style       | The default styles for nodes. |
| node-styles      | A map of classes to node styles. Used by the `class` field on nodes. |
| link-style       | The default styles for links. |
| link-styles      | A map of classes to link styles. Used by the `class` field on links. |
| node-label-style | Styles for node labels. |
| link-label-style | Styles for link labels. |
| link-color-scale | The color scale used to map link values to colors. |

The default config is:

    {
      "min-node-sep": 5,
      "node-style": {
        "size": 20,
        "style": {
          "stroke-width": 4,
          "stroke":       "#000000",
          "fill":         "#ffffff"
        }
      },
      "link-style": {
        "size":   10,
        "radius": 10,
        "style": {
          "stroke-width": 0,
          "fill":         "#808080"
        }
      },
      "node-label-style": {
        "size":        16,
        "font-family": "sans-serif",
        "color":       "#000000"
      },
      "link-label-style": {
        "size":          8,
        "font-family":   "monospace",
        "color":         "#000000",
        "background":    "#ffffff",
        "border":        "#000000",
        "opacity":       0.9,
        "border-radius": 3,
        "width":         28
      },
      "link-color-scale": < See Below >
    }
    
Run `make-map -dumpconf` to see the default config

## NodeStyle & LinkStyle

`NodeStyle` and `LinkStyle` have the following common fields:

    {
      "opacity": float,
      "fill": Color,
      "stroke": Color,
      "stroke-width": float,
    }

| Field        | Description |
| ---:         | :---        |
| opacity      | A value between 0 and 1 of how opaque the object is. Default: 1 (opaque) |
| fill         | The color used to fill the object |
| stroke       | The color used for the outline of the object |
| stroke-width | The width of the outline of the object |

`NodeStyle` has the following additional fields

    {
      "size": float
    }
    
| Field        | Description |
| ---:         | :---        |
| size         | The size of the node. Specifically diameter of the node. |

`LinkStyle` has the following additional fields

    {
      "size": float,
      "radius": float
    }
    
| Field        | Description |
| ---:         | :---        |
| size         | The size of the link. Specifically the width of link. |
| radius       | The corner radius of the rendered link. Set to 0 to disable rounded corners. |

## NodeLabelStyle & LinkLabelStyle

`NodeLabelStyle` and `LinkLabelStyle` have the following common fields:

    {
      "size": float,
      "color": Color,
      "font-family": string
    }

| Field        | Description |
| ---:         | :---        |
| size         | Size of the text |
| color        | Color of the text |
| font-family  | The font family/face used |

`LinkLabelStyle` has the following additional fields

    {
      "background-color": Color,
      "border-color": Color,
      "border-radius": float,
      "width": float,
      "opacity": float
    }

| Field            | Description |
| ---:             | :---        |
| background-color | The background color for the link labels |
| border-color     | The color of the border around the labels |
| border-radius    | The corner radius of the the border. Set to 0 for square corners. |
| width            | The total width of the label. This is fixed for all link labels. |
| opacity          | The opacity of the label's background |

## Color & ColorScale

`Color` is a string containing the hex-coded RGB value for a color, i.e. `"#abcdef"`.

`ColorScale` describes a mapping between values and colors. Colors are assigned specific values
and interpolated between to provide the colors for other values.
There are two formats for `ColorScale`

    {
      "space": "rgb"/"hsl",
      "colors": [[float, Color]]
    }
    // or
    [[float, Color]]

| Field            | Description |
| ---:             | :---        |
| space            | The color space to interpolate in. Defaults to HSL. |
| colors           | A list of value/color pairs. This is also the second format. |

### Heat Scale

The default color scale is a "heat" scale with the following description:

    {
      "space": "hsl",
      "colors: [
        [  0, "#1d4877"],
        [0.1, "#1b8a5a"],
        [0.5, "#fbb01f"],
        [0.7, "#f68838"],
        [0.9, "#ee3e32"]
      ]
    }
