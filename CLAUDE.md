# CLITasks

CLI task management tool with Markdown storage, usable by humans and AI agents.

## Tech Stack

- TypeScript, Node.js
- `commander` for CLI, `blessed`/`blessed-contrib` for TUI
- Markdown files as storage (no database)

## Build & Dev

```bash
npm install        # Install dependencies
npm run build      # Compile TypeScript (tsc)
npm run dev -- <cmd>       # Run CLI in dev mode (tsx)
npm run dev:tui            # Run TUI in dev mode
npm start -- <cmd>         # Run built CLI
```

## Project Structure

- `src/` — TypeScript source
  - `cli.ts` — CLI entry point (commander)
  - `tui.ts` — Terminal UI (blessed)
  - `board.ts` — Task board logic
  - `storage.ts` — Markdown file parser/writer
  - `export.ts` — Export functionality
  - `types.ts` — Type definitions

## Git Rules

- **Never add `Co-Authored-By` trailers to commit messages.**
- Write concise commit messages focused on the "why."
