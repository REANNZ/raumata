SVG Target Details
==================

SVG is an XML-based format that, importantly, is embeddable in HTML.
As such, the structure of the SVG document is important for users that
want to add interactivity or other dynamic elements to the map.

## Map Document Structure

Starting from `Renderer.RenderTopology`, the structure is:

``` svg
<g id="#topology">
  <g id="links">
    <!-- Links -->
  </g>
  <g id="nodes">
    <!-- Nodes -->
  </g>
</g>
```

### Link Structure

Ignoring style information, the structure of a link in the map is:

``` svg
<g id="L-<LinkId>" class="link">
  <g class="link-segment" data-from="<NodeId>" data-to="<NodeId>">
    <path d="<data>" />
    <g class="link-label">
      <rect class="link-label-box" />
      <text class="link-label-text">LABEL</text>
    </g>
  </g>
  <g class="link-segment" data-from="<NodeId>" data-to="<NodeId>">
    <path d="<data>" />
    <g class="link-label">
      <rect class="link-label-box" />
      <text class="link-label-text">LABEL</text>
    </g>
  </g>
</g>
```

If there is no label, then the link label group will be ommited.

### Node Structure

Ignoring style information, the structure of a node in the map is:

``` svg
<g id="N-<NodeId>" data-node="<NodeId>">
  <circle class="node" />
  <text class="node-label-text">LABEL</text>
</g>
```
