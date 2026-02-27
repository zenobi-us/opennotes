# Changelog

## [0.2.0](https://github.com/zenobi-us/jot/compare/v0.1.0...v0.2.0) (2026-02-22)


### ⚠ BREAKING CHANGES

* **memory:** Step 3 has been updated to reflect that release-please is already configured

### Features

* **docs:** add basic getting started guide for non-power users ([2ebe37c](https://github.com/zenobi-us/jot/commit/2ebe37cf617f9b1609c1600c58a5489bf676613c))
* implement pi-opennotes extension phase 2 ([c75c891](https://github.com/zenobi-us/jot/commit/c75c891340170535a10e377886f7b25048de7800))
* **notebook:** improve migrate list status output ([#18](https://github.com/zenobi-us/jot/issues/18)) ([d537a76](https://github.com/zenobi-us/jot/commit/d537a765793e4024d32768555c5481a9089c128c))
* **notes-view:** add save/delete flags for notebook custom views ([c9a9542](https://github.com/zenobi-us/jot/commit/c9a95423fdb03d7fa0929da8bdaeb7ec3714d74c))
* remove DuckDB, migrate to Bleve search with semantic search enhancement ([#17](https://github.com/zenobi-us/jot/issues/17)) ([c81925e](https://github.com/zenobi-us/jot/commit/c81925e487dcbc103b8dc4ff41f621e7b6345238))
* **search,views:** add OR DSL support and view execution overrides ([354f290](https://github.com/zenobi-us/jot/commit/354f29075727c5b340a6dfd6c2cebf52db064d7e))
* use "." as root when initializing in existing directories ([80e767a](https://github.com/zenobi-us/jot/commit/80e767ac484000c35811fe7d455df2c606472092))
* **view:** implement GROUP BY, DISTINCT, OFFSET, HAVING, aggregates, and template enhancements (task-3d477ab8) ([#16](https://github.com/zenobi-us/jot/issues/16)) ([bb83706](https://github.com/zenobi-us/jot/commit/bb8370667ee78608cec32b91e47acf7c243f3028))


### Bug Fixes

* force next release to 0.2.0 ([25e02f6](https://github.com/zenobi-us/jot/commit/25e02f678540ea32ccd121e953b87aedc6a3a8e4))
* **notes-view:** tighten mutating flag validation errors ([39ea50c](https://github.com/zenobi-us/jot/commit/39ea50cbc6b472e1f4ae9bed0ed4db2c24056c1e))
* **view:** honor view precedence when listing ([0a92ef9](https://github.com/zenobi-us/jot/commit/0a92ef9b03b3a00c29ea1c030fbcc44250e0cbb3))
* **view:** validate and resolve parameters in deterministic order ([8cf2be8](https://github.com/zenobi-us/jot/commit/8cf2be82dcc59a663cc1d91bc6b73a5b7e52c304))


### Documentation

* **memory:** update GitHub Actions CI/CD task with current repo state ([7c19e94](https://github.com/zenobi-us/jot/commit/7c19e941692e8e3833f0363f187320ab7daed41e))

## [0.1.0](https://github.com/zenobi-us/jot/compare/v0.0.3...v0.1.0) (2026-01-25)


### Features

* Advanced Search System with Views & Enhanced Note Creation ([#9](https://github.com/zenobi-us/jot/issues/9)) ([b517454](https://github.com/zenobi-us/jot/commit/b517454d518bd04609035ebbf9defbbf41c784f2))
* **notebook:** implement JOT_NOTEBOOK envvar and fix resolution priority ([#14](https://github.com/zenobi-us/jot/issues/14)) ([349683c](https://github.com/zenobi-us/jot/commit/349683cfa4143c8c2605d87b03596f2ce9e1053a))
* **schema:** add JSON schema for notebook configuration ([#10](https://github.com/zenobi-us/jot/issues/10)) ([38f4fbe](https://github.com/zenobi-us/jot/commit/38f4fbe1ad75c77da38dceedda83e6dca66f8a7d))


### Bug Fixes

* **docs:** convert notebook discovery flowchart from Mermaid to D2 ([#7](https://github.com/zenobi-us/jot/issues/7)) ([6c56012](https://github.com/zenobi-us/jot/commit/6c5601237b03c1e54d79840cd059174a633e512e))
* **views:** correct DuckDB metadata schema for all built-in views ([#11](https://github.com/zenobi-us/jot/issues/11)) ([5da5fe9](https://github.com/zenobi-us/jot/commit/5da5fe9e96f0d63dce8faeb55da9df760c7f600e))

## [Unreleased]

### Features

#### Semantic Search (Optional Enhancement) ✨

**Vector-based search for finding notes by meaning, not just keywords**

Jot now supports semantic search that understands concepts and paraphrases, supplementing the existing full-text search.

**Command**:
```bash
jot notes search semantic [query] [--mode hybrid|keyword|semantic] [--explain]
```

**Search Modes**:
- **Hybrid (default)**: Combines keyword + semantic retrieval using RRF merge
- **Keyword**: Fast full-text search via Bleve index
- **Semantic**: Meaning-based search via vector embeddings

**Features**:
- Hybrid retrieval with deterministic RRF (Reciprocal Rank Fusion) merge
- Boolean filters work across all modes (`--and`, `--or`, `--not`)
- Explain mode shows match type and reasoning per result
- Graceful fallback to keyword-only when semantic backend unavailable
- Sub-200ms latency for typical notebook sizes

**Examples**:
```bash
# Hybrid search (default)
jot notes search semantic "project planning discussions"

# With filters
jot notes search semantic "architecture" --and data.tag=design --not data.status=archived

# Explain output
jot notes search semantic "workflow" --explain
```

**When to use**:
- **Regular search**: Exact keywords, specific terms, quick lookups
- **Semantic search**: Conceptual queries, paraphrases, exploratory search

**Documentation**: [Semantic Search Guide](docs/semantic-search-guide.md)

**Implementation**: chromem-go vector backend with automatic indexing lifecycle

---

### BREAKING CHANGES

#### Search Engine Migration: DuckDB → Bleve

**SQL interface removed from NoteService**

Jot now uses Bleve full-text search exclusively. The SQL interface methods have been removed:
- `ExecuteSQLSafe()` - removed
- `Query()` - removed
- `--sql` flag in `notes search` command - removed

**Migration Guide**:

**Before** (SQL):
```bash
jot notes search --sql "SELECT * FROM read_markdown('**/*.md') WHERE content LIKE '%meeting%'"
```

**After** (Bleve):
```bash
# Simple text search
jot notes search "meeting"

# DSL boolean query
jot notes search "data.tag:work NOT data.status:archived"
```

**Why?**
- Eliminates DuckDB dependency (32MB+ binary size reduction)
- Faster indexing and search for typical use cases
- Simpler deployment (no C++ dependencies)
- Better cross-platform compatibility

**Affected Users**:
- Custom SQL queries no longer supported
- Power users should wait for Phase 5.3 (link graph) or use external tools

See: Epic [epic-f661c068-remove-duckdb-alternative-search.md](.memory/epic-f661c068-remove-duckdb-alternative-search.md)

**Known Issues**:
- Tag filtering (`--and data.tag=value`) not working - under investigation (array indexing)
- Fuzzy search needs tuning for optimal fuzziness distance
- Link queries (`links-to`, `linked-by`) deferred to Phase 5.3 (requires graph index)

See: [research-55e8a9f3-phase54-known-issues.md](.memory/research-55e8a9f3-phase54-known-issues.md)

### Features

#### Views System ✨
**Complete reusable query presets for common workflows**

- **6 Built-in Views**: `today`, `recent`, `kanban`, `untagged`, `orphans`, `broken-links`
- **Custom Views**: Define views in global config (`~/.config/jot/config.json`) or notebook config (`.jot.json`)
- **Parameter System**: Runtime parameters with validation (string, list, date, bool types)
- **Template Variables**: Dynamic date/time values (`{{today}}`, `{{yesterday}}`, `{{this_week}}`, `{{this_month}}`, `{{now}}`)
- **Multiple Outputs**: List, table, and JSON formats for all views
- **Special Executors**: Graph analysis for orphans and broken-links detection
- **Performance**: Sub-millisecond query generation, <50ms total execution

**Command**:
```bash
jot notes view <name> [--param key=value] [--format list|table|json]
jot notes view --list  # Show available views
```

**Examples**:
```bash
# View today's notes
jot notes view today

# Kanban board with custom statuses
jot notes view kanban --param status=todo,in-progress,done

# Find orphaned notes
jot notes view orphans --param definition=no-incoming

# Export to JSON for automation
jot notes view recent --format json | jq '.[].path'
```

**Documentation**:
- User Guide: `docs/views-guide.md`
- Examples: `docs/views-examples.md`
- API Reference: `docs/views-api.md`

**Implementation Details**:
- Core data structures: `ViewDefinition`, `ViewParameter`, `ViewQuery`, `ViewCondition`
- Service: `internal/services/view.go` - ViewService with 6 built-in views
- Special executors: `internal/services/view_special.go` - Orphans and broken-links detection
- CLI command: `cmd/notes_view.go`
- Configuration integration: 3-tier precedence (notebook > global > built-in)
- Security: Field/operator whitelist + parameterized queries
- Test coverage: 59 new tests (100% ViewService and SpecialViewExecutor)

---

## [0.0.3](https://github.com/zenobi-us/jot/compare/0.1.0-next.1...v0.0.3) (2026-01-20)


### ⚠ BREAKING CHANGES

* Remove Node.js dependencies and package.json

### Bug Fixes

* adjust for prerelease tag offset in version computation ([14aa55a](https://github.com/zenobi-us/jot/commit/14aa55af7b1e394c7985461a05ba3d1217bf4f60))
* **ci:** correct release-please outputs printing in workflow ([35b45a6](https://github.com/zenobi-us/jot/commit/35b45a68533bc9326ad7e82ae6d694b83d604676))
* force tag creation for release workflow ([933add0](https://github.com/zenobi-us/jot/commit/933add04b2300e1c1315dc3b6a6fd8ec474112e8))
* ignore coverage files ([5a9a27e](https://github.com/zenobi-us/jot/commit/5a9a27e91246d8bbe5f03d000e8d1159b809c0a8))
* resolve all bats test failures and security issues ([#6](https://github.com/zenobi-us/jot/issues/6)) ([9353f1c](https://github.com/zenobi-us/jot/commit/9353f1c70fe38cd8cb9759dc0b0f53be76c448f4))


### Code Refactoring

* migrate from Node.js to Go-native version management ([51846b0](https://github.com/zenobi-us/jot/commit/51846b0b167a00295605761b880f2f3c694b9873))

## [0.0.2](https://github.com/zenobi-us/jot/compare/v0.0.1...v0.0.2) (2026-01-17)


### Bug Fixes

* **publish:** fetch git tags in checkout action ([5d6af78](https://github.com/zenobi-us/jot/commit/5d6af785086ebf729603f28baf38badd3cb24adb))

## 0.0.1 (2026-01-17)


### Features

* **cli:** add --sql flag to notes search command ([780acdd](https://github.com/zenobi-us/jot/commit/780acdd9dcbc321d2fba805f0c633e54fc6abe56))
* **core:** add wiki notebook and notes management system ([c3ae87f](https://github.com/zenobi-us/jot/commit/c3ae87fe787ba792b1acfb44aef57101dca362bc))
* **db:** add GetReadOnlyDB() method for safe query execution ([bffdf90](https://github.com/zenobi-us/jot/commit/bffdf901ec691aaafa76cc1e281eaca2e4141f6a))
* **display:** add RenderSQLResults() for table formatting ([a4dcc91](https://github.com/zenobi-us/jot/commit/a4dcc91b02ef03855ed5144947cb786f5c6db36d))
* Go rewrite with comprehensive testing and CI/CD ([#1](https://github.com/zenobi-us/jot/issues/1)) ([62d21ab](https://github.com/zenobi-us/jot/commit/62d21abb6c746ad8b609ac6755ce0145e741ff11))
* **init:** add init command and refactor ConfigService ([e34ed99](https://github.com/zenobi-us/jot/commit/e34ed99275f5d39f089183847764e297ef23519d))
* **sql:** add NoteService.ExecuteSQLSafe() for query orchestration ([5b9d9e2](https://github.com/zenobi-us/jot/commit/5b9d9e259e58df4f86654b633335c0719423b6f7))
* **sql:** add ValidateSQL() for safe query execution ([74ce4af](https://github.com/zenobi-us/jot/commit/74ce4afeffb987056412fe3dcfd15e274f2accff))


### Bug Fixes

* correct GoReleaser configuration for jot build ([744dc50](https://github.com/zenobi-us/jot/commit/744dc505011a09667949e47e9c66d2395552dacb))
* **publish:** add GITHUB_TOKEN and clarify release target ([4079f6b](https://github.com/zenobi-us/jot/commit/4079f6be8e5460ab9ebe4a7b4419e8f460a38824))
* **publish:** create git tag before GoReleaser runs ([bf8fdbf](https://github.com/zenobi-us/jot/commit/bf8fdbfe4b82e9f92bb8e570f1386c0d9f9f3500))
* **publish:** use git tags instead of calculating prerelease versions ([a144cb7](https://github.com/zenobi-us/jot/commit/a144cb764b8ea3d5c453372130e80767aea67821))
* **types:** resolve all TypeScript type errors ([bf69925](https://github.com/zenobi-us/jot/commit/bf69925146ca05743ac600378643fd4c2d05ed5f))

## Changelog

All notable changes to this project will be documented here by Release Please.
