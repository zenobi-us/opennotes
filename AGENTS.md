# AGENTS.md

## Build & Test Commands

Always run commands from the project root using `mise run <command>`.

If you think you need to use `bun <command>`, stop and get help first.

- **Build**: `mise run build` 
- **Test**: `mise run test`
- **Single Test**: `mise run test BackgroundTask.test.ts` (use file glob pattern)
- **Watch Mode**: `mise run test --watch`
- **Lint**: `mise run lint` (eslint)
- **Fix Lint**: `mise run lint:fix` (eslint --fix)
- **Format**: `mise run format` (prettier)

## Code Style Guidelines

### Imports & Module System

- Use ES6 `import`/`export` syntax (module: "ESNext", type: "module")
- Group imports: external libraries first, then internal modules
- Use explicit file extensions (`.ts`) for internal imports

### Formatting (Prettier)

- **Single quotes** (`singleQuote: true`)
- **Line width**: 100 characters
- **Tab width**: 2 spaces
- **Trailing commas**: ES5 (no trailing commas in function parameters)
- **Semicolons**: enabled

### TypeScript & Naming

- **NeverNesters**: avoid deeply nested structures. Always exit early.
- **Strict mode**: enforced (`"strict": true`)
- **Classes**: PascalCase (e.g., `BackgroundTask`, `BackgroundTaskManager`)
- **Methods/properties**: camelCase
- **Status strings**: use union types (e.g., `'pending' | 'running' | 'completed' | 'failed' | 'cancelled'`)
- **Explicit types**: prefer explicit type annotations over inference
- **Return types**: optional (not required but recommended for public methods)

### Error Handling

- Check error type before accessing error properties: `error instanceof Error ? error.toString() : String(error)`
- Log errors with `[ERROR]` prefix for consistency
- Always provide error context when recording output

### Linting Rules

- `@typescript-eslint/no-explicit-any`: warn (avoid `any` type)
- `no-console`: error (minimize console logs)
- `prettier/prettier`: error (formatting violations are errors)

## Testing

- Framework: **vitest** with `describe` & `it` blocks
- Style: Descriptive nested test cases with clear expectations
- Assertion library: `expect()` (vitest)

## Project Context

- **Type**: ES Module package for OpenCode plugin system
- **Target**: Bun runtime, ES2021+
- **Purpose**: Background task execution and lifecycle management

## Architecture

opennotes is currently a cli tool for managing notes.

### Data Flow

1. Cli instance initialises
2. Commands are registered
3. User invokes a command via CLI
4. Root interceptor initialises context.store with the following services:
  - ConfigService
  - DbService
  - NotebookService
  - NotesService
5. Command handler routes to the appropriate command handler.

### Components

#### ConfigService

Manages application configuration settings, including loading, saving, and validating config data.

Only user global config settings here. Per-notebook settings go in NotebookService.

#### DbService

We use this when we have a notebook opened. It gives us SQL query layer on top of a filesystem of markdown files.

#### NotebookService

Abstracts all the notebook-level operations, including 

