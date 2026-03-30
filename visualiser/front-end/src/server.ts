// This server is used to test the visualiser.
// Run with: npm run serve -- <path-to-json>

import fs from "node:fs";
import http, { IncomingMessage, ServerResponse } from "node:http";
import path from "node:path";

const port = Number(process.env.PORT || 3000);
const dataArg = process.argv[2];

if (!dataArg) {
  console.error("Usage: npm run serve -- <path-to-json>");
  process.exit(1);
}

const dataPath = path.resolve(process.cwd(), dataArg);

if (!fs.existsSync(dataPath)) {
  console.error(`file not found: ${dataPath}`);
  process.exit(1);
}

const rootDir = process.cwd();

function sendFile(res: ServerResponse, filePath: string, contentType: string): void {
  fs.readFile(filePath, (err, content) => {
    if (err) {
      res.writeHead(404, { "Content-Type": "text/plain; charset=utf-8" });
      res.end("Not found");
      return;
    }

    res.writeHead(200, { "Content-Type": contentType });
    res.end(content);
  });
}

function route(req: IncomingMessage, res: ServerResponse): void {
  const url = req.url || "/";

  if (url === "/data") {
    sendFile(res, dataPath, "application/json; charset=utf-8");
    return;
  }

  if (url === "/" || url === "/index.html") {
    sendFile(res, path.join(rootDir, "public", "index.html"), "text/html; charset=utf-8");
    return;
  }

  if (url === "/index.css") {
    sendFile(res, path.join(rootDir, "public", "index.css"), "text/css; charset=utf-8");
    return;
  }

  if (url === "/dist/index.js") {
    sendFile(res, path.join(rootDir, "dist", "index.js"), "application/javascript; charset=utf-8");
    return;
  }

  res.writeHead(404, { "Content-Type": "text/plain; charset=utf-8" });
  res.end("Not found");
}

http.createServer(route).listen(port, () => {
  console.log(`Serving ${dataPath} at http://localhost:${port}`);
});
