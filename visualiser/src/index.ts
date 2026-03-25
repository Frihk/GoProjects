import ForceGraph, { GraphData, LinkObject, NodeObject } from "force-graph";

interface AntRoom extends NodeObject {
  group?: "start" | "end" | "room";
}

interface AntFarmData extends GraphData {
  nodes: AntRoom[];
  links: LinkObject[];
}

// Raw fx/fy values in test data are grid units; scale them for visible spacing.
const scale = 60;

const elem = document.getElementById("graph-container") as HTMLElement;

const Graph = new ForceGraph<AntRoom>(elem)
  .nodeId("id")
  .enableNodeDrag(false)
  .linkWidth(2)
  .linkColor(() => "#666")
  .nodeCanvasObject((node, ctx, globalScale) => {
    const fontSize = 16 / globalScale;
    ctx.font = `${fontSize}px Sans-Serif`;

    let color = "#4da6ff"; // Default room (Blue)
    if (node.group === "start") color = "#4caf50"; // ##start (Green)
    if (node.group === "end") color = "#f44336"; // ##end (Red)

    // Draw the room (circle)
    ctx.beginPath();
    ctx.arc(node.x!, node.y!, 10, 0, 2 * Math.PI, false);
    ctx.fillStyle = color;
    ctx.fill();

    // Draw the room name (text)
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.fillStyle = "white";
    ctx.fillText(node.id as string, node.x!, node.y!);
  });

fetch("/data")
  .then((response) => response.json())
  .then((data: AntFarmData) => {
    const scaledData: AntFarmData = {
      ...data,
      nodes: data.nodes.map((node) => ({
        ...node,
        x: typeof node.fx === "number" ? node.fx * scale : node.x,
        y: typeof node.fy === "number" ? node.fy * scale : node.y,
        fx: typeof node.fx === "number" ? node.fx * scale : node.fx,
        fy: typeof node.fy === "number" ? node.fy * scale : node.fy,
      })),
    };

    Graph.graphData(scaledData);
    setTimeout(() => {
      Graph.zoomToFit(200, 100); // 200ms transition, 100px padding
    }, 100);
  })
  .catch((error) => console.error("Error loading ant farm data:", error));
