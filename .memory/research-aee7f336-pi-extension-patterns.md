---
id: aee7f336
title: Pi Extension Patterns Research
type: "research"
created_at: 2026-01-28T23:25:00+10:30
updated_at: 2026-01-28T23:25:00+10:30
status: completed
epic_id: 1f41631e
---

# Pi Extension Patterns Research

## Research Questions

1. How do pi extensions register custom tools?
2. What is the best structure for a pi package with tools?
3. How do other extensions handle CLI tool integration?
4. What are the truncation best practices for tool output?

## Summary

Pi extensions are TypeScript modules that export a default function receiving `ExtensionAPI`. They can register tools, commands, event handlers, and custom renderers. Key findings:

- Tools use TypeBox schemas for parameters
- `StringEnum` must be used instead of `Type.Union` for Google API compatibility
- Output truncation uses built-in utilities (`truncateHead`, `truncateTail`)
- State persistence uses `pi.appendEntry()` with session restoration

## Findings

### 1. Tool Registration Pattern

```typescript
import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";
import { Type } from "@sinclair/typebox";
import { StringEnum } from "@mariozechner/pi-ai";

export default function (pi: ExtensionAPI) {
  pi.registerTool({
    name: "my_tool",
    label: "My Tool",
    description: "Tool description for LLM",
    parameters: Type.Object({
      action: StringEnum(["list", "add"] as const),
      query: Type.Optional(Type.String()),
    }),
    async execute(toolCallId, params, onUpdate, ctx, signal) {
      // Check for cancellation
      if (signal?.aborted) {
        return { content: [{ type: "text", text: "Cancelled" }] };
      }
      
      // Execute command
      const result = await pi.exec("command", ["args"], { signal });
      
      return {
        content: [{ type: "text", text: result.stdout }],
        details: { exitCode: result.code },
      };
    },
  });
}
```

### 2. Package Structure for npm

```
pkgs/pi-opennotes/
├── package.json          # npm package with "pi" manifest
├── src/
│   └── index.ts         # Main extension entry
├── tsconfig.json
└── README.md
```

**package.json structure:**
```json
{
  "name": "@zenobi-us/pi-opennotes",
  "version": "0.1.0",
  "keywords": ["pi-package"],
  "main": "src/index.ts",
  "peerDependencies": {
    "@mariozechner/pi-coding-agent": "*",
    "@sinclair/typebox": "*"
  },
  "pi": {
    "extensions": ["./src/index.ts"]
  }
}
```

### 3. CLI Integration Patterns (from ssh.ts example)

```typescript
// Execute external commands
const result = await pi.exec("opennotes", ["notes", "search", query], { 
  signal,
  timeout: 30000 
});

// Handle errors
if (result.code !== 0) {
  return {
    content: [{ type: "text", text: `Error: ${result.stderr}` }],
    isError: true,
  };
}
```

### 4. Output Truncation Pattern

```typescript
import {
  truncateHead,
  DEFAULT_MAX_BYTES,  // 50KB
  DEFAULT_MAX_LINES,  // 2000
  formatSize,
} from "@mariozechner/pi-coding-agent";

const truncation = truncateHead(output, {
  maxLines: DEFAULT_MAX_LINES,
  maxBytes: DEFAULT_MAX_BYTES,
});

let result = truncation.content;

if (truncation.truncated) {
  // Write full output to temp file
  result += `\n\n[Output truncated: ${truncation.outputLines} lines]`;
}
```

### 5. Session State Persistence

```typescript
// Persist state
pi.appendEntry("opennotes-config", { notebook: "/path/to/notebook" });

// Restore on session start
pi.on("session_start", async (_event, ctx) => {
  for (const entry of ctx.sessionManager.getBranch()) {
    if (entry.type === "custom" && entry.customType === "opennotes-config") {
      // Restore state
    }
  }
});
```

### 6. Proposed Tool Set for pi-opennotes

| Tool | Description | Parameters |
|------|-------------|------------|
| `opennotes_search` | Search notes with SQL or full-text | `query: string`, `notebook?: string`, `limit?: number` |
| `opennotes_list` | List notes in a notebook | `notebook?: string`, `limit?: number` |
| `opennotes_get` | Get a specific note by path | `path: string`, `notebook?: string` |
| `opennotes_create` | Create a new note | `title: string`, `template?: string`, `notebook?: string` |
| `opennotes_notebooks` | List available notebooks | (none) |
| `opennotes_views` | List or execute views | `action: "list" | "execute"`, `view?: string` |

## References

1. [Pi Extensions Documentation](~/.local/share/mise/installs/npm-mariozechner-pi-coding-agent/0.50.1/lib/node_modules/@mariozechner/pi-coding-agent/docs/extensions.md)
2. [Pi Packages Documentation](~/.local/share/mise/installs/npm-mariozechner-pi-coding-agent/0.50.1/lib/node_modules/@mariozechner/pi-coding-agent/docs/packages.md)
3. [tools.ts Example](~/.local/share/mise/installs/npm-mariozechner-pi-coding-agent/0.50.1/lib/node_modules/@mariozechner/pi-coding-agent/examples/extensions/tools.ts)
4. [ssh.ts Example](~/.local/share/mise/installs/npm-mariozechner-pi-coding-agent/0.50.1/lib/node_modules/@mariozechner/pi-coding-agent/examples/extensions/ssh.ts)
