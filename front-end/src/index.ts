import ForceGraph, {
    LinkObject,
    NodeObject,
} from "force-graph";

// ---------------------------------------------------------
// Types
// ---------------------------------------------------------

interface AntRoom extends NodeObject {
    id: string;
    group?: "start" | "end" | "room";
    fx?: number;
    fy?: number;
    x?: number;
    y?: number;
}

interface Move {
    antId: number;
    roomName: string;
}

interface Step {
    moves: Move[];
}

interface AntFarmData {
    ants: number;
    start: string;
    end: string;
    nodes: AntRoom[];
    links: LinkObject<AntRoom>[];
    steps: Step[];
}

interface AntState {
    id: number;
    from: AntRoom;
    to: AntRoom;
    progress: number;
    arrived: boolean;
}

// Declare lucide global (loaded via <script> tag in HTML).
declare const lucide: { createIcons: () => void };

// ---------------------------------------------------------
// Constants
// ---------------------------------------------------------

/** Scale factor applied to raw grid coordinates for visible spacing. */
const COORD_SCALE = 50;

/** How much progress (0‑1) an ant makes per animation frame. */
const ANIMATION_SPEED = 0.02;

// ---------------------------------------------------------
// DOM Elements
// ---------------------------------------------------------

const turnDisplay = document.getElementById("turn-display")!;
const statTotal = document.getElementById("stat-total")!;
const statArrived = document.getElementById("stat-arrived")!;
const progressBar = document.getElementById("progress-bar") as HTMLElement;
const btnPlay = document.getElementById("btn-play")!;
const btnStep = document.getElementById("btn-step")!;
const btnReset = document.getElementById("btn-reset")!;
const graphContainer = document.getElementById("graph-container")!;

// ---------------------------------------------------------
// Global State
// ---------------------------------------------------------

let allSteps: Step[] = [];
let currentTurnIdx = -1;
const ants = new Map<number, AntState>();
let startRoom: AntRoom | null = null;
let endRoom: AntRoom | null = null;
let totalAntsCount = 0;
let isPlaying = false;
let isRendering = false;

// ---------------------------------------------------------
// Render loop – keeps the canvas alive while ants are moving
// ---------------------------------------------------------

function ensureRendering(): void {
    if (isRendering) return;
    isRendering = true;

    (function loop() {
        let moving = false;
        ants.forEach((a) => {
            if (a.progress < 1) moving = true;
        });

        if (isPlaying || moving) {
            // Nudge zoom by an imperceptible amount to force a canvas repaint
            Graph.zoom(Graph.zoom() + (Math.random() > 0.5 ? 1e-10 : -1e-10));
            requestAnimationFrame(loop);
        } else {
            isRendering = false;
        }
    })();
}

// ---------------------------------------------------------
// Graph Setup
// ---------------------------------------------------------

const Graph = new ForceGraph<AntRoom>(graphContainer)
    .nodeId("id")
    .enableNodeDrag(false)
    .linkWidth(2)
    .linkColor(() => "rgba(255, 255, 255, 0.1)")
    .nodeCanvasObject((node: AntRoom, ctx: CanvasRenderingContext2D, globalScale: number) => {
        const fontSize = 14 / globalScale;
        ctx.font = `600 ${fontSize}px 'Outfit', sans-serif`;

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

        const x = node.x!;
        const y = node.y!;

        // Glow
        ctx.beginPath();
        const gradient = ctx.createRadialGradient(x, y, 4, x, y, 14);
        gradient.addColorStop(0, glow);
        gradient.addColorStop(1, "rgba(0,0,0,0)");
        ctx.fillStyle = gradient;
        ctx.arc(x, y, 14, 0, 2 * Math.PI);
        ctx.fill();

        // Circle
        ctx.beginPath();
        ctx.arc(x, y, 7, 0, 2 * Math.PI, false);
        ctx.fillStyle = color;
        ctx.fill();
        ctx.strokeStyle = "rgba(255,255,255,0.2)";
        ctx.lineWidth = 1;
        ctx.stroke();

        // Label
        ctx.textAlign = "center";
        ctx.textBaseline = "middle";
        ctx.fillStyle = "rgba(255, 255, 255, 0.8)";
        ctx.fillText(node.id, x, y + 16);
    })
    .onRenderFramePost((ctx: CanvasRenderingContext2D) => {
        let arrivedCount = 0;

        ants.forEach((ant) => {
            const fromX = ant.from.x;
            const fromY = ant.from.y;
            const toX = ant.to.x;
            const toY = ant.to.y;

            if (ant.arrived) arrivedCount++;

            if (
                fromX !== undefined &&
                fromY !== undefined &&
                toX !== undefined &&
                toY !== undefined
            ) {
                const x = fromX + (toX - fromX) * ant.progress;
                const y = fromY + (toY - fromY) * ant.progress;

                // Ant glow
                ctx.beginPath();
                const antGlow = ctx.createRadialGradient(x, y, 1, x, y, 6);
                antGlow.addColorStop(0, "#fbbf24");
                antGlow.addColorStop(1, "rgba(251, 191, 36, 0)");
                ctx.fillStyle = antGlow;
                ctx.arc(x, y, 6, 0, 2 * Math.PI);
                ctx.fill();

                // Ant dot
                ctx.beginPath();
                ctx.arc(x, y, 3, 0, 2 * Math.PI, false);
                ctx.fillStyle = "#fff";
                ctx.fill();

                // Animate
                if (isPlaying || ant.progress < 1) {
                    ant.progress = Math.min(1, ant.progress + ANIMATION_SPEED);
                    if (ant.progress >= 1 && ant.to.group === "end") {
                        ant.arrived = true;
                    }
                }
            }
        });

        // Update HUD
        statArrived.textContent = String(arrivedCount);
        const pct = totalAntsCount > 0 ? (arrivedCount / totalAntsCount) * 100 : 0;
        progressBar.style.width = `${pct}%`;

        // Auto-advance turns
        if (isPlaying && Array.from(ants.values()).every((a) => a.progress >= 1)) {
            if (currentTurnIdx < allSteps.length - 1) {
                nextTurn();
            } else {
                setTimeout(() => setPlaying(false), 100);
            }
        }

        // Stop when all ants have arrived
        if (
            totalAntsCount > 0 &&
            arrivedCount === totalAntsCount &&
            Array.from(ants.values()).every((a) => a.progress >= 1)
        ) {
            setPlaying(false);
        }
    });

// ---------------------------------------------------------
// Play / Pause
// ---------------------------------------------------------

function setPlaying(val: boolean): void {
    isPlaying = val;
    const icon = btnPlay.querySelector("i");
    if (icon) {
        icon.setAttribute("data-lucide", isPlaying ? "pause" : "play");
        lucide.createIcons();
    }
    if (isPlaying) ensureRendering();
}

// ---------------------------------------------------------
// Simulation Logic
// ---------------------------------------------------------

function nextTurn(): void {
    if (currentTurnIdx >= allSteps.length - 1) return;
    currentTurnIdx++;
    turnDisplay.textContent = String(currentTurnIdx + 1);

    const step = allSteps[currentTurnIdx];
    const graphNodes = Graph.graphData().nodes as AntRoom[];

    step.moves.forEach((move) => {
        const targetRoom = graphNodes.find((n) => n.id === move.roomName);
        if (!targetRoom) return;

        let ant = ants.get(move.antId);
        if (!ant) {
            ant = {
                id: move.antId,
                from: startRoom!,
                to: targetRoom,
                progress: 0,
                arrived: false,
            };
            ants.set(move.antId, ant);
        } else {
            ant.from = ant.to;
            ant.to = targetRoom;
            ant.progress = 0;
        }
    });

    ensureRendering();
}

function resetSimulation(): void {
    currentTurnIdx = -1;
    ants.clear();
    turnDisplay.textContent = "0";
    setPlaying(false);
    statArrived.textContent = "0";
    progressBar.style.width = "0%";
}

// ---------------------------------------------------------
// Data Fetching (JSON from Go server)
// ---------------------------------------------------------

fetch("/data")
    .then((res) => res.json())
    .then((data: AntFarmData) => {
        // Scale coordinates for visible spacing
        const scaledNodes: AntRoom[] = data.nodes.map((node) => ({
            ...node,
            fx: typeof node.fx === "number" ? node.fx * COORD_SCALE : node.fx,
            fy: typeof node.fy === "number" ? node.fy * COORD_SCALE : node.fy,
        }));

        allSteps = data.steps ?? [];
        totalAntsCount = data.ants ?? 0;
        statTotal.textContent = String(totalAntsCount);

        startRoom = scaledNodes.find((n) => n.group === "start") ?? null;
        endRoom = scaledNodes.find((n) => n.group === "end") ?? null;

        Graph.graphData({ nodes: scaledNodes, links: data.links });
        setTimeout(() => Graph.zoomToFit(400, 100), 500);
    })
    .catch((err) => console.error("Error loading ant farm data:", err));

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
