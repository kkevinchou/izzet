# Izzet Log Viewer

Vite + React browser app for interleaving structured JSONL logs by timestamp.

Run it from this directory:

```powershell
npm install
npm run dev
```

Then open the Vite URL. The app auto-loads `logs/client.log` and `logs/server.log` when they exist. You can also choose `logs/client.log`, `logs/server.log`, `logs/garbage.log`, and/or `logs/app.log` manually. The app reads files through browser file picker or the folder picker if your browser supports it.

Expected structured log shape is JSON per line. The viewer recognizes `time`, `timestamp`, `ts`, `level`, `msg`, `message`, `source`, `tags`, `tag`, `labels`, and `label`. Tags can be arrays, comma/space separated strings, or objects such as:

```json
{"time":"11:00:24.064","level":"INFO","msg":"read stale output","tags":["net","prediction"],"cf":666,"gcf":284}
```
