import { defineConfig, type Plugin, type PreviewServer, type ViteDevServer } from "vite";
import react from "@vitejs/plugin-react";
import { readFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";

const DEFAULT_LOG_FILES = [
  { name: "client.log", path: fileURLToPath(new URL("../../logs/client.log", import.meta.url)) },
  { name: "server.log", path: fileURLToPath(new URL("../../logs/server.log", import.meta.url)) },
];

export default defineConfig({
  plugins: [react(), defaultLogsPlugin()],
  server: {
    fs: {
      allow: [fileURLToPath(new URL("../..", import.meta.url))],
    },
  },
});

function defaultLogsPlugin(): Plugin {
  return {
    name: "default-logs",
    configureServer(server) {
      registerDefaultLogRoute(server);
    },
    configurePreviewServer(server) {
      registerDefaultLogRoute(server);
    },
  };
}

function registerDefaultLogRoute(server: ViteDevServer | PreviewServer) {
  server.middlewares.use("/default-logs", async (_request, response) => {
    const results = await Promise.all(
      DEFAULT_LOG_FILES.map(async (file) => {
        try {
          return { name: file.name, text: await readFile(file.path, "utf8") };
        } catch {
          return null;
        }
      }),
    );
    const sources = results.filter((source): source is { name: string; text: string } => source !== null);

    response.setHeader("Content-Type", "application/json");
    response.statusCode = sources.length ? 200 : 404;
    response.end(JSON.stringify({ sources }));
  });
}
