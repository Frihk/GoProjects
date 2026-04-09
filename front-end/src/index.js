// ---------------------------------------------------------
// UI Elements
// ---------------------------------------------------------
const turnDisplay = document.getElementById("turn-display");
const statTotal = document.getElementById("stat-total");
const statArrived = document.getElementById("stat-arrived");
const progressBar = document.getElementById("progress-bar");

const btnPlay = document.getElementById("btn-play");
const btnStep = document.getElementById("btn-step");
const btnReset = document.getElementById("btn-reset");

// ---------------------------------------------------------
// Global State
// ---------------------------------------------------------
let allSteps = [];
let currentTurnIdx = -1;
let ants = new Map();
let startRoom = null;
let endRoom = null;
let totalAntsCount = 0;
let isPlaying = false;
let animationSpeed = 0.02;

let isRendering = false;

// ---------------------------------------------------------
// Render Force
// ---------------------------------------------------------
function ensureRendering() {
    if (!isRendering) {
        isRendering = true;
        (function loop() {
            let moving = false;
            ants.forEach(a => { if (a.progress < 1) moving = true; });
            
            if (isPlaying || moving) {
                // Apply an imperceptible zoom tweak to force the canvas to repaint
                // This keeps animation active even when layout physics is asleep
                Graph.zoom(Graph.zoom() + (Math.random() > 0.5 ? 1e-10 : -1e-10));
                requestAnimationFrame(loop);
            } else {
                isRendering = false;
            }
        })();
    }
}

// ---------------------------------------------------------
// Graph Setup
// ---------------------------------------------------------
const elem = document.getElementById("graph-container");
const Graph = ForceGraph()(elem)
  .nodeId("id")
  .enableNodeDrag(false)
  .linkWidth(2)
  .linkColor(() => "rgba(255, 255, 255, 0.1)")
  .nodeCanvasObject((node, ctx, globalScale) => {
    const fontSize = 14 / globalScale;
    ctx.font = `600 ${fontSize}px 'Outfit', Sans-Serif`;

    let color = "#3b82f6"; 
    let glow = "rgba(59, 130, 246, 0.3)";
    
    if (node.group === "start") {
      color = "#10b981";
      glow = "rgba(16, 185, 129, 0.4)";
    }
    if (node.group === "end") {
      color = "#f43f5e";
      glow = "rgba(244, 63, 94, 0.4)";
    }

    // Draw Glow
    ctx.beginPath();
    const gradient = ctx.createRadialGradient(node.x, node.y, 4, node.x, node.y, 14);
    gradient.addColorStop(0, glow);
    gradient.addColorStop(1, "rgba(0,0,0,0)");
    ctx.fillStyle = gradient;
    ctx.arc(node.x, node.y, 14, 0, 2 * Math.PI);
    ctx.fill();

    // Draw the room (circle)
    ctx.beginPath();
    ctx.arc(node.x, node.y, 7, 0, 2 * Math.PI, false);
    ctx.fillStyle = color;
    ctx.fill();
    ctx.strokeStyle = "rgba(255,255,255,0.2)";
    ctx.lineWidth = 1;
    ctx.stroke();

    // Draw the room name (text)
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.fillStyle = "rgba(255, 255, 255, 0.8)";
    ctx.fillText(node.id, node.x, node.y + 16);
  })
  .onRenderFramePost((ctx) => {
    let arrivedCount = 0;
    
    ants.forEach((ant) => {
      const from = ant.from;
      const to = ant.to;

      if (ant.arrived) arrivedCount++;

      if (from.x !== undefined && from.y !== undefined && to.x !== undefined && to.y !== undefined) {
        const x = from.x + (to.x - from.x) * ant.progress;
        const y = from.y + (to.y - from.y) * ant.progress;

        ctx.beginPath();
        const antGlow = ctx.createRadialGradient(x, y, 1, x, y, 6);
        antGlow.addColorStop(0, "#fbbf24"); 
        antGlow.addColorStop(1, "rgba(251, 191, 36, 0)");
        ctx.fillStyle = antGlow;
        ctx.arc(x, y, 6, 0, 2 * Math.PI);
        ctx.fill();

        ctx.beginPath();
        ctx.arc(x, y, 3, 0, 2 * Math.PI, false);
        ctx.fillStyle = "#fff";
        ctx.fill();

        if (isPlaying || ant.progress < 1) {
             ant.progress = Math.min(1, ant.progress + animationSpeed);
             if (ant.progress >= 1 && ant.to.group === "end") {
                 ant.arrived = true;
             }
        }
      }
    });

    if (statArrived) statArrived.textContent = arrivedCount;
    if (progressBar) {
        const pct = totalAntsCount > 0 ? (arrivedCount / totalAntsCount) * 100 : 0;
        progressBar.style.width = `${pct}%`;
    }

    if (isPlaying && Array.from(ants.values()).every(a => a.progress >= 1)) {
        if (currentTurnIdx < allSteps.length - 1) {
            nextTurn();
        } else {
            // Give a tiny delay for final ant to smooth in before stopping auto-play
            setTimeout(() => setPlaying(false), 100);
        }
    }

    if (totalAntsCount > 0 && arrivedCount === totalAntsCount && Array.from(ants.values()).every(a => a.progress >= 1)) {
        setPlaying(false);
    }
  });

function setPlaying(val) {
    isPlaying = val;
    const icon = btnPlay.querySelector('i');
    if (isPlaying) {
        icon.setAttribute('data-lucide', 'pause');
    } else {
        icon.setAttribute('data-lucide', 'play');
    }
    lucide.createIcons();
    if (isPlaying) ensureRendering();
}

// ---------------------------------------------------------
// Parser
// ---------------------------------------------------------
function parseLemIn(text) {
  const lines = text.split("\n").map(l => l.trim()).filter(l => l !== "");
  const nodes = [];
  const links = [];
  const steps = [];
  let antsCount = 0;
  let moveBlockStarted = false;

  let pendingGroup = "";

  lines.forEach((line) => {
    if (line.startsWith("##")) {
        if (line === "##start") pendingGroup = "start";
        if (line === "##end") pendingGroup = "end";
        return;
    }
    if (line.startsWith("#")) return;
    if (line.startsWith("L")) {
      moveBlockStarted = true;
      const moves = line.split(" ").map(m => {
        const parts = m.substring(1).split("-");
        return { antId: parseInt(parts[0]), roomName: parts[1] };
      });
      steps.push({ moves });
      return;
    }
    if (moveBlockStarted) return;

    const fields = line.split(" ");
    if (fields.length === 3) {
      const node = { 
          id: fields[0], 
          fx: parseInt(fields[1]) * 50, 
          fy: parseInt(fields[2]) * 50,
          group: pendingGroup || "room"
      };
      nodes.push(node);
      pendingGroup = "";
    } else if (line.includes("-")) {
      const parts = line.split("-");
      links.push({ source: parts[0], target: parts[1] });
    } else if (!isNaN(parseInt(line)) && antsCount === 0) {
      antsCount = parseInt(line);
    }
  });

  return { nodes, links, steps, antsCount };
}

// ---------------------------------------------------------
// Simulation Logic
// ---------------------------------------------------------
function nextTurn() {
    if (currentTurnIdx >= allSteps.length - 1) return;
    currentTurnIdx++;
    turnDisplay.innerHTML = currentTurnIdx + 1;

    const step = allSteps[currentTurnIdx];
    step.moves.forEach(move => {
        let ant = ants.get(move.antId);
        const targetRoom = Graph.graphData().nodes.find(n => n.id === move.roomName);

        if (!targetRoom) return;

        if (!ant) {
            ant = {
                id: move.antId,
                from: startRoom,
                to: targetRoom,
                progress: 0,
                arrived: false
            };
            ants.set(move.antId, ant);
        } else {
            ant.from = ant.to;
            ant.to = targetRoom;
            ant.progress = 0;
        }
    });
}

function resetSimulation() {
    currentTurnIdx = -1;
    ants.clear();
    turnDisplay.innerHTML = "0";
    setPlaying(false);
    
    if (statArrived) statArrived.textContent = "0";
    if (progressBar) progressBar.style.width = "0%";
}

// ---------------------------------------------------------
// Data Fetching
// ---------------------------------------------------------
fetch("/api/raw")
  .then(res => res.text())
  .then(text => {
    if (!text) return;
    const data = parseLemIn(text);
    allSteps = data.steps;
    totalAntsCount = data.antsCount;
    if (statTotal) statTotal.textContent = totalAntsCount;
    startRoom = data.nodes.find(n => n.group === "start") || null;
    endRoom = data.nodes.find(n => n.group === "end") || null;

    Graph.graphData({ nodes: data.nodes, links: data.links });
    setTimeout(() => Graph.zoomToFit(400, 100), 500);
  });

// ---------------------------------------------------------
// Event Listeners
// ---------------------------------------------------------
btnPlay.onclick = () => {
    const newVal = !isPlaying;
    setPlaying(newVal);
    if (isPlaying && currentTurnIdx === -1 && allSteps.length > 0) {
        nextTurn();
    }
};

btnStep.onclick = () => {
    setPlaying(false);
    nextTurn();
    ensureRendering();
};

btnReset.onclick = () => {
    resetSimulation();
};
