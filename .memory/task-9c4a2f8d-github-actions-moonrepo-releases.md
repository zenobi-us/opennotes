---
id: 9c4a2f8d
title: GitHub Actions CI/CD with moonrepo, release-please, and GoReleaser
type: "task"
created_at: 2026-01-29T17:33:00+10:30
updated_at: 2026-01-29T18:54:00+10:30
status: todo
epic_id: null
phase_id: null
assigned_to: null
---

# GitHub Actions CI/CD with moonrepo and release-please

## Objective

Update GitHub Actions workflows to integrate moonrepo's affected command detection and release-please manifest mode for independent, safe package releases in a monorepo structure.

## Important Clarifications

### Workflow File Targets

**CRITICAL CORRECTION**: We ARE modifying `.github/workflows/release.yml` - it's OUR release workflow, not moonrepo's

| Workflow File | Purpose | Action |
|--------------|---------|--------|
| `release.yml` | **Our release-please automation** | ✅ MODIFY (remove hardcoded release-type) |
| `publish.yml` | **Our package publishing workflow** | ✅ MODIFY (add moonrepo integration) |
| `pr.yml` | PR quality gate (optional) | ✅ CREATE |

**Why this matters**: 
- `release.yml` already exists and uses release-please-action@v4
- It's OUR workflow for automated releases, NOT moonrepo's internal tooling
- No need to create a separate `release-please.yml` - we already have `release.yml`
- `publish.yml` handles actual package publishing (triggered by release.yml)

### What publish.yml Actually Does

**Current Understanding** (needs implementation):

```yaml
# .github/workflows/publish.yml
name: Publish Packages

# Triggered by:
on:
  push:
    branches: [main]  # After release-please PR merge
  push:
    tags:
      - 'go-api-v*'   # For GoReleaser

# What it does:
jobs:
  # 1. Test affected packages using moonrepo
  affected-tests:
    run: moon ci --base main
  
  # 2. Detect which packages changed
  detect-affected:
    run: moon query projects --affected --json
  
  # 3. Publish Node/Bun packages via mise
  publish-node:
    run: mise run node:publish ${{ affected_packages }}
  
  # 4. Publish Go binaries via GoReleaser (if tag)
  publish-go:
    uses: goreleaser/goreleaser-action@v5
```

**Key Integration Points**:
1. **moonrepo affected detection** → determines what to test/publish
2. **mise variadic tasks** → publish specific packages
3. **release-please** → creates version bump PRs (separate workflow or integrated)
4. **GoReleaser** → builds Go binaries on tag creation

### Mise Task Reorganization

As part of this work, we need to reorganize mise tasks to support multiple package ecosystems:

**Current Structure**:
```
.mise/tasks/publish
```

**New Structure**:
```
.mise/tasks/
├── node/
│   └── publish    # Node/Bun package publishing
├── go/
│   └── publish    # Go binary publishing (future)
```

**Rationale**:
- Namespaces tasks by language/runtime
- Allows different publish workflows for different package types
- Keeps tasks organized as the monorepo grows
- Enables calling specific publish tasks: `mise run node:publish`

**Mise Variadic Arguments**:

Mise tasks can be implemented as **bash scripts** that accept variadic arguments using standard bash `$@`:

```bash
#!/usr/bin/env bash
# .mise/tasks/node/publish

PACKAGES=("$@")

if [ ${#PACKAGES[@]} -eq 0 ]; then
  # Publish all packages when no arguments provided
  pnpm publish -r
else
  # Publish specific packages
  for pkg in "${PACKAGES[@]}"; do
    pnpm publish --filter "$pkg"
  done
fi
```

This enables:
- `mise run node:publish` - Publish all packages (no args)
- `mise run node:publish @zenobi-us/opennotes` - Publish specific package
- `mise run node:publish pkg1 pkg2` - Publish multiple packages

This variadic capability integrates with moonrepo's affected detection, allowing selective publishing based on what changed.

**Why bash scripts instead of TOML**:
- Simpler syntax for conditionals and loops
- Direct access to bash features (`$@`, arrays, etc.)
- No templating syntax needed
- Mise executes bash scripts as tasks automatically

## Context

This task was deferred during Phase 4 (Polish) of the Go migration work. The original requirement was to "Update CI/CD for Go builds" but this has expanded to include:

1. Supporting multiple packages (Go and TypeScript/Bun) in the same repository
2. Using moonrepo's dependency graph to determine what needs testing
3. Using release-please for automatic version bumping and changelog generation
4. Only releasing packages that have actual changes

### Original Deferred Task Location

From `.memory/archive/01-migrate-to-golang/task-1k2l3m4n-polish.md`:
- Section 4.7 Build Configuration
- Reason deferred: Needed post-merge integration after Go rewrite was production-ready

## Research Findings

### Strategy: Implicit Detection + Dependency Graph (Recommended)

After evaluating multiple approaches, the recommended strategy is **"Strategy 3"** which combines:
- **release-please** with implicit detection (only bump versions for packages with commits in their scope)
- **moonrepo** as the dependency graph enforcer (ensures dependent packages are tested)

This approach provides:
- ✅ Clean, independent version bumps (no linked versions)
- ✅ Safety through dependency-aware testing
- ✅ Prevents releases if dependent packages break
- ✅ Simple configuration and maintenance

### moonrepo Integration

**Purpose**: Dependency graph enforcement and affected test detection

**Key Commands**:
- `moon ci --base main` - Run tests for affected projects based on git diff
- This ensures that when `go-api` changes, `web-app` tests also run (if web-app depends on go-api)

**Configuration** (`moon.yml`):
```yaml
# In apps/web-app/moon.yml
dependsOn:
  - services/go-api
```

**Benefits**:
- Automatically discovers what needs testing based on changes
- Respects dependency relationships
- Prevents breaking changes from being released

### release-please Configuration

**Purpose**: Automatic version bumping and changelog generation for packages with actual changes

**Mode**: Manifest mode with separate package definitions

**Key Configuration** (`release-please-config.json`):
```json
{
  "packages": {
    "services/go-api": {
      "component": "go-api",
      "release-type": "go",
      "changelog-path": "CHANGELOG.md"
    },
    "apps/web-app": {
      "component": "web-app", 
      "release-type": "node",
      "changelog-path": "CHANGELOG.md"
    }
  }
}
```

**Important**: Do NOT use `linked-versions` - this would force all packages to bump together, defeating the purpose of independent releases.

**Behavior**:
- Scans commit messages following Conventional Commits
- Only bumps versions for packages with commits in their scope (e.g., `fix(go-api): ...`)
- Creates separate PRs for each package that needs releasing
- Updates version files automatically:
  - Node/Bun: `package.json`
  - Go: `version.go` or similar (needs setup)

### Workflow Architecture

**CRITICAL**: Understanding which workflows we modify vs which we leave alone:

#### 1. `.github/workflows/release.yml` - ALREADY EXISTS (Uses release-please)

**Purpose**: Automated release management with release-please

**Current State**: ✅ Already configured and working

**What it does**:
- Runs release-please-action on push to main
- Creates release PRs with version bumps and changelog
- Creates GitHub releases when release PRs are merged
- Dispatches publish events to trigger package publishing

**Existing Configuration**:
```yaml
- uses: google-github-actions/release-please-action@v4
  id: release-please
  with:
    token: ${{ secrets.GITHUB_TOKEN }}
    release-type: go  # ⚠️ NEEDS UPDATE (see Step 3)
    skip-github-pull-request: false
```

**What Needs Fixing**:
- Remove hardcoded `release-type: go` (conflicts with per-package config)
- Let `release-please-config.json` control release types per package

**Why This File Exists**: This is OUR workflow for release automation, NOT moonrepo's internal releases

#### 2. `.github/workflows/publish.yml` - MODIFY THIS

**Purpose**: Package publishing using moonrepo + mise + release-please

**What it does**:
- Detects affected packages using `moon query projects --affected`
- Runs quality gates (tests for affected packages)
- Calls `mise run node:publish` with variadic package arguments
- Publishes only packages that have changes and pass tests

**Triggers**:
- On repository_dispatch (triggered by release.yml after release creation)
- On successful tag creation (for Go binaries via GoReleaser)

**Key Integration**:
```yaml
- name: Get affected packages
  id: affected
  run: moon query projects --affected --json

- name: Publish affected Node packages
  if: steps.affected.outputs.node_packages != ''
  run: mise run node:publish ${{ steps.affected.outputs.node_packages }}
```

#### 3. `.github/workflows/pr.yml` - Optional Quality Gate

**Purpose**: PR validation before merge

**What it does**:
- Runs `moon ci --base main` to test affected packages
- Ensures dependent packages don't break
- Blocks merge if tests fail

**Integration**:
```yaml
- name: Run affected tests
  run: moon ci --base ${{ github.base_ref }}
```


## Current Repository State

Before starting implementation, here's what already exists:

### Version Tracking (Go Package)

**Version Variables Location**: `cmd/root.go` (lines 12-16)
```go
var (
	Version   string
	BuildDate string
	GitCommit string
	GitBranch string
)
```

**Version Command**: `cmd/version.go` (already complete)
- Displays version, build date, git commit, and branch
- Integrated with cobra CLI via `opennotes version`
- Uses the Version variables from cmd/root.go

**Current release-please Configuration**:
```json
{
  "packages": {
    ".": {
      "extra-files": ["main.go"]  // ⚠️ Missing "cmd/root.go"
    }
  }
}
```

**What's Missing**:
- `cmd/root.go` not listed in extra-files (needs to be added)
- release-please cannot currently update the Version variable
- Single-package mode (needs conversion to monorepo mode)

### Existing Files

1. **release-please-config.json**
   - Status: ✅ Exists, needs conversion
   - Current mode: Single-package (root ".")
   - Release type: "simple"
   - Extra files: ["main.go"] ⚠️ Missing "cmd/root.go"

2. **.release-please-manifest.json**
   - Status: ✅ Exists, needs update
   - Current version: "0.1.0" for root package
   - Needs: Multiple package entries with separate versions

3. **.github/workflows/release.yml**
   - Status: ✅ Exists, needs modification
   - Purpose: Runs release-please on push to main
   - Current issue: Hardcoded `release-type: go` overrides per-package config
   - Action needed: Remove hardcoded release-type, let config control it
   - Jobs:
     - `process`: Runs release-please-action@v4
     - `dispatch-publish`: Dispatches publish-package events

4. **.goreleaser.yaml**
   - Status: ✅ Exists
   - Purpose: GoReleaser configuration for Go binary builds
   - Integration: Triggered by tag creation from release-please

### Files That Don't Exist Yet

1. **.github/workflows/pr.yml** - PR quality gate (to be created)
2. **.mise/tasks/node/publish** - Namespaced publish task (needs migration from `.mise/tasks/publish`)
3. **moon.yml** - moonrepo configuration (needs verification/creation)
4. **Package-level moon.yml files** - Dependency declarations (to be created)

### Files That Already Exist

1. **.github/workflows/release.yml** - ✅ Uses release-please-action@v4 (needs config update)
2. **.github/workflows/publish.yml** - ✅ Exists (needs moonrepo integration)
3. **cmd/root.go** - ✅ Contains Version, BuildDate, GitCommit, GitBranch variables
4. **cmd/version.go** - ✅ Complete version command implementation

### What Needs to Change

| Component | Current State | Target State | Action Required |
|-----------|---------------|--------------|-----------------|
| release-please-config.json | Single-package mode (".") | Multi-package monorepo | Convert configuration |
| .release-please-manifest.json | Single entry (".": "0.1.0") | Multiple package entries | Update manifest |
| .github/workflows/release.yml | Hardcoded `release-type: go` | Config-driven per-package | Remove hardcoded type |
| cmd/root.go extra-files | Not listed in config | Included in extra-files | Add "cmd/root.go" to config |
| .mise/tasks/publish | Single flat file | Namespaced (node/publish) | Reorganize structure |
| .github/workflows/publish.yml | Basic structure | Full moonrepo integration | Add affected detection |
| .github/workflows/pr.yml | Doesn't exist | Moonrepo quality gate | Create new workflow |

## Steps

### 1. Reorganize Mise Tasks

**Goal**: Create namespaced task structure for different package types

**IMPORTANT**: Mise tasks should remain as **bash scripts**, NOT TOML files. Mise can execute bash scripts directly as tasks.

- [ ] Create directory structure:
  ```bash
  mkdir -p .mise/tasks/node
  mkdir -p .mise/tasks/go
  ```
- [ ] Move existing publish task:
  ```bash
  git mv .mise/tasks/publish .mise/tasks/node/publish
  ```
- [ ] Update `.mise/tasks/node/publish` to accept variadic arguments (keep as bash script):
  ```bash
  #!/usr/bin/env bash
  # .mise/tasks/node/publish
  # Description: Publish Node/Bun packages
  # Usage: mise run node:publish [@zenobi-us/pkg1] [@zenobi-us/pkg2]
  
  # Accept package names as arguments using $@
  PACKAGES=("$@")
  
  # If no packages specified, publish all
  if [ ${#PACKAGES[@]} -eq 0 ]; then
    echo "Publishing all packages..."
    pnpm publish -r
  else
    # Publish only specified packages
    echo "Publishing specific packages: ${PACKAGES[*]}"
    for pkg in "${PACKAGES[@]}"; do
      echo "Publishing $pkg..."
      pnpm publish --filter "$pkg"
    done
  fi
  ```
- [ ] Ensure the bash script is executable:
  ```bash
  chmod +x .mise/tasks/node/publish
  ```
- [ ] Test task invocation:
  ```bash
  mise run node:publish --help
  mise run node:publish --dry-run
  mise run node:publish @zenobi-us/opennotes
  mise run node:publish @zenobi-us/pkg1 @zenobi-us/pkg2
  ```
- [ ] Document new task structure in project docs

**Why Bash Scripts Instead of TOML**:
- ✅ Simpler syntax for shell operations
- ✅ Direct use of bash features (`$@`, arrays, conditionals)
- ✅ No need for templating syntax
- ✅ Mise executes bash scripts as tasks automatically
- ✅ More flexible for complex logic

**References**:
- [Mise Tasks Documentation](https://mise.jdx.dev/tasks/)
- Bash variadic arguments via `$@` enable selective publishing: `mise run node:publish pkg1 pkg2`

### 2. Setup moonrepo Configuration

- [ ] Verify `moon.yml` exists in project root
- [ ] Define dependencies between packages in individual `moon.yml` files
  ```yaml
  # Example: apps/web-app/moon.yml
  dependsOn:
    - services/go-api
  ```
- [ ] Test locally: `moon ci --base main`

### 3. Convert release-please to Monorepo Mode

**CURRENT STATE**: release-please is already configured, but in **single-package mode**

**Existing Files**:
- ✅ `release-please-config.json` - Currently configured for single root package
- ✅ `.release-please-manifest.json` - Current version: "0.1.0"
- ✅ `.github/workflows/release.yml` - Already uses release-please-action@v4

**Current Configuration** (`release-please-config.json`):
```json
{
  "packages": {
    ".": {
      "extra-files": ["main.go"]
    }
  },
  "$schema": "https://raw.githubusercontent.com/googleapis/release-please/main/schemas/config.json",
  "include-v-in-tag": true,
  "include-component-in-tag": false,
  "versioning": "prerelease",
  "prerelease": true,
  "bump-minor-pre-major": true,
  "release-type": "simple"
}
```

**What This Means**:
- Single package at root (".") - treats entire repo as one package
- `release-type: "simple"` - basic versioning without language-specific features
- Only updates `main.go` as extra file
- NOT configured for monorepo with multiple independent packages

**What Needs to Change**:
This is a **conversion from single-package to multi-package monorepo mode**, not a setup from scratch.

**Conversion Steps**:

- [ ] Backup current configuration:
  ```bash
  cp release-please-config.json release-please-config.json.bak
  cp .release-please-manifest.json .release-please-manifest.json.bak
  ```

- [ ] Convert `release-please-config.json` to multi-package manifest mode:
  ```json
  {
    "packages": {
      "services/go-api": {
        "component": "go-api",
        "release-type": "go",
        "changelog-path": "CHANGELOG.md",
        "extra-files": ["cmd/root.go"]
      },
      "apps/web-app": {
        "component": "web-app",
        "release-type": "node",
        "changelog-path": "CHANGELOG.md"
      },
      "pkgs/pi-opennotes": {
        "component": "pi-opennotes",
        "release-type": "node",
        "changelog-path": "CHANGELOG.md"
      }
    },
    "$schema": "https://raw.githubusercontent.com/googleapis/release-please/main/schemas/config.json",
    "include-v-in-tag": true,
    "include-component-in-tag": true,
    "bump-minor-pre-major": true
  }
  ```

- [ ] Update `.release-please-manifest.json` with entries for each package:
  ```json
  {
    "services/go-api": "0.1.0",
    "apps/web-app": "0.1.0",
    "pkgs/pi-opennotes": "0.1.0"
  }
  ```

- [ ] **CRITICAL**: Update `.github/workflows/release.yml` to remove hardcoded `release-type`:
  - Current: `release-type: go` (hardcoded at workflow level)
  - Problem: This overrides per-package release types in config
  - Fix: Remove `release-type` from workflow, let config file control it
  
  **Change from**:
  ```yaml
  - uses: google-github-actions/release-please-action@v4
    id: release-please
    with:
      token: ${{ secrets.GITHUB_TOKEN }}
      release-type: go  # ❌ REMOVE THIS LINE
      skip-github-pull-request: false
  ```
  
  **To**:
  ```yaml
  - uses: google-github-actions/release-please-action@v4
    id: release-please
    with:
      token: ${{ secrets.GITHUB_TOKEN }}
      # release-type controlled by release-please-config.json per-package
      skip-github-pull-request: false
  ```

- [ ] Test conversion with dry-run (if release-please CLI supports it):
  ```bash
  # Install release-please CLI
  npm install -g release-please
  
  # Validate new configuration
  release-please manifest-pr --dry-run
  ```

- [ ] Create test commit with conventional commit format to verify detection:
  ```bash
  git commit -m "feat(go-api): test monorepo detection"
  ```

- [ ] Push and verify release-please creates PR with correct package detection

- [ ] Verify Conventional Commits usage across project (existing requirement)

**Key Changes**:
1. **Single root package** → **Multiple packages with paths**
2. **`release-type: "simple"`** → **Per-package release types** (`go`, `node`)
3. **`include-component-in-tag: false`** → **`include-component-in-tag: true`** (enables tags like `go-api-v1.0.0`)
4. **Single manifest entry** → **Multiple package entries** with independent versions
5. **Workflow hardcoded release-type** → **Config-driven per-package types**

**Why This Conversion Matters**:
- Enables independent versioning per package (not all packages bump together)
- Allows different release strategies per package (Go vs Node)
- Creates component-specific tags (e.g., `go-api-v1.0.0` instead of just `v1.0.0`)
- Supports Conventional Commits scoped to packages: `feat(go-api): ...`

### 4. Configure Go Version Tracking

**CURRENT STATE**: Version variables already exist in `cmd/root.go`

**Existing Implementation**:
```go
// cmd/root.go lines 12-16
var (
	Version   string
	BuildDate string
	GitCommit string
	GitBranch string
)
```

**Version Command**: Already implemented in `cmd/version.go`
```go
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Print detailed version information for OpenNotes including build metadata",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("OpenNotes %s\n", Version)
		if BuildDate != "unknown" {
			fmt.Printf("Built: %s\n", BuildDate)
		}
		if GitCommit != "unknown" {
			fmt.Printf("Commit: %s\n", GitCommit)
		}
		if GitBranch != "unknown" {
			fmt.Printf("Branch: %s\n", GitBranch)
		}
	},
}
```

**What Needs to Be Done**:
- [ ] Verify `cmd/root.go` contains Version variable declaration (✅ Confirmed above)
- [ ] Update `release-please-config.json` to include `cmd/root.go` in extra-files for go-api package:
  ```json
  "services/go-api": {
    "component": "go-api",
    "release-type": "go",
    "changelog-path": "CHANGELOG.md",
    "extra-files": ["cmd/root.go"]
  }
  ```
  Note: This will be part of the monorepo conversion in Step 3
- [ ] Ensure release-please can detect and update the Version string in cmd/root.go
- [ ] Test version update workflow with a test commit
- [ ] Verify `opennotes version` command displays updated version after release

**Why cmd/root.go Instead of version.go**:
- Version variables are declared in `cmd/root.go`, not `cmd/version.go`
- `cmd/version.go` only contains the command definition that *uses* the variables
- release-please needs to update the file where variables are *declared*

### 5. Update GitHub Actions Workflows

**CRITICAL FILE MAPPING**:

| File | Purpose | Action |
|------|---------|--------|
| `.github/workflows/publish.yml` | Package publishing (our custom workflow) | **MODIFY** |
| `.github/workflows/pr.yml` | PR quality gate (optional) | **CREATE/MODIFY** |
| `.github/workflows/release.yml` | Moonrepo's own releases | **DO NOT TOUCH** |
| `.github/workflows/release-please.yml` | Release-please automation (optional) | **CREATE IF NEEDED** |

**5.1 Update or Create `pr.yml` (PR Quality Gate)**

**Purpose**: Validate PRs before merge using moonrepo affected detection

**Action**: Create or modify `.github/workflows/pr.yml`

Checklist:
- [ ] Add moonrepo setup step
- [ ] Replace test commands with `moon ci --base ${{ github.base_ref }}`
- [ ] Ensure full git history is fetched (`fetch-depth: 0`)
- [ ] Add Go toolchain setup for Go packages
- [ ] Add Bun setup for TypeScript packages
- [ ] Test with sample PR to verify affected detection works

**Example Structure**:
```yaml
name: PR Quality Gate

on:
  pull_request:
    branches: [main]

jobs:
  test-affected:
    name: Test Affected Packages
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      
      - uses: moonrepo/setup-toolchain@v0
      
      - name: Run tests for affected projects
        run: moon ci --base ${{ github.base_ref }}
```

**5.2 Modify `publish.yml` (Package Publishing Workflow)**

**IMPORTANT**: Modify `.github/workflows/publish.yml`, NOT `release.yml`

**Purpose**: Publish packages using moonrepo affected detection + mise tasks

**Action**: Update `.github/workflows/publish.yml` with new structure

**CRITICAL: Extract Inline Logic to Testable Script**

The moonrepo query and jq filtering logic should NOT be embedded directly in the YAML workflow. Instead, create a testable bash script.

**Why Extract to Script**:
- ✅ Testable locally without pushing to GitHub
- ✅ Easier to debug (add echo statements, test standalone)
- ✅ Can add unit tests for filtering logic
- ✅ Reusable across multiple workflows
- ✅ Clear separation of concerns
- ✅ Iterate quickly on filtering logic

Checklist:
- [ ] **Create `.github/scripts/` directory**:
  ```bash
  mkdir -p .github/scripts
  ```

- [ ] **Create `.github/scripts/get-affected-pkgs.sh`** script:
  ```bash
  #!/usr/bin/env bash
  # .github/scripts/get-affected-pkgs.sh
  # Purpose: Get affected packages from moonrepo for CI/CD pipelines
  # Usage: ./get-affected-pkgs.sh [base-branch] [package-type]
  # Example: ./get-affected-pkgs.sh main node
  #          ./get-affected-pkgs.sh main go
  
  set -euo pipefail
  
  # Accept arguments with defaults
  BASE_BRANCH="${1:-main}"
  PKG_TYPE="${2:-node}"
  DRY_RUN="${DRY_RUN:-false}"
  
  # Debug output if dry-run mode
  if [[ "$DRY_RUN" == "true" ]]; then
    echo "DEBUG: Base branch: $BASE_BRANCH" >&2
    echo "DEBUG: Package type: $PKG_TYPE" >&2
  fi
  
  # Get affected projects from moonrepo
  AFFECTED=$(moon query projects --affected --json --base "$BASE_BRANCH")
  
  # Debug: Show raw moonrepo output
  if [[ "$DRY_RUN" == "true" ]]; then
    echo "DEBUG: Affected projects JSON:" >&2
    echo "$AFFECTED" | jq '.' >&2
  fi
  
  # Filter by package type and extract package names
  # Note: Adjust jq filter based on your moonrepo project structure
  # Common patterns:
  #   - .language == "node" or .language == "go"
  #   - .tags[] | contains("node") or .tags[] | contains("go")
  #   - .type == "application" or .type == "library"
  PACKAGES=$(echo "$AFFECTED" | jq -r \
    --arg type "$PKG_TYPE" \
    '.[] | select(.language == $type) | .id' | tr '\n' ' ')
  
  # Trim trailing whitespace
  PACKAGES=$(echo "$PACKAGES" | xargs)
  
  # Debug: Show filtered packages
  if [[ "$DRY_RUN" == "true" ]]; then
    echo "DEBUG: Filtered packages: $PACKAGES" >&2
  fi
  
  # Output space-separated package names (for mise to consume)
  echo "$PACKAGES"
  ```

- [ ] **Make script executable**:
  ```bash
  chmod +x .github/scripts/get-affected-pkgs.sh
  ```

- [ ] **Test script locally before committing**:
  ```bash
  # Test with dry-run mode to see debug output
  DRY_RUN=true bash .github/scripts/get-affected-pkgs.sh main node
  
  # Test actual output (what CI will receive)
  bash .github/scripts/get-affected-pkgs.sh main node
  
  # Test with different base branches
  bash .github/scripts/get-affected-pkgs.sh develop node
  
  # Test go package detection
  bash .github/scripts/get-affected-pkgs.sh main go
  
  # Test with explicit environment variable
  DRY_RUN=false bash .github/scripts/get-affected-pkgs.sh main node
  ```

- [ ] **Update workflow to use the script** (in `.github/workflows/publish.yml`):
  ```yaml
  - name: Get affected Node packages
    id: affected-node
    run: |
      PKGS=$(.github/scripts/get-affected-pkgs.sh ${{ github.base_ref || 'main' }} node)
      echo "packages=$PKGS" >> $GITHUB_OUTPUT
  
  - name: Get affected Go packages
    id: affected-go
    run: |
      PKGS=$(.github/scripts/get-affected-pkgs.sh ${{ github.base_ref || 'main' }} go)
      echo "packages=$PKGS" >> $GITHUB_OUTPUT
  ```

- [ ] **Add quality gate job** (runs moonrepo tests on affected packages)
- [ ] **Add publish-node job** that calls `mise run node:publish` with script output:
  ```yaml
  - name: Publish affected Node packages
    if: steps.affected-node.outputs.packages != ''
    run: mise run node:publish ${{ steps.affected-node.outputs.packages }}
  ```
- [ ] **Add publish-go job** triggered by tag creation (for GoReleaser)
- [ ] Ensure jobs have proper dependencies (quality-gate → detect → publish)
- [ ] Add conditional execution (`if:` checks) to prevent unnecessary runs
- [ ] Configure npm authentication for Node package publishing
- [ ] Test workflow with dry-run before actual publishing

**Script Benefits**:
1. **Local Testing**: Run `bash .github/scripts/get-affected-pkgs.sh main node` to test filtering
2. **Debugging**: Set `DRY_RUN=true` to see verbose output
3. **Iteration**: Modify jq filter and test immediately
4. **Reusability**: Other workflows can use the same script
5. **Clarity**: Logic is in one place, not scattered across YAML
6. **Testability**: Can write unit tests for the script
7. **Version Control**: Script changes are clearly visible in git history

**5.3 Release-Please Integration (Already Complete)**

**CURRENT STATE**: ✅ Release-please already integrated in `.github/workflows/release.yml`

**Existing Architecture**:
- **release.yml** - Runs release-please-action on push to main
- **publish.yml** - Triggered by repository_dispatch events from release.yml

This is **Option B** (separate workflows) and is already implemented:
- ✅ Clean separation of concerns
- ✅ Easier to understand workflow logic  
- ✅ Release-please runs independently
- ✅ File already exists and works

**What Just Needs Fixing**:
- [ ] Remove hardcoded `release-type: go` from release.yml (Step 3)
- [ ] Let per-package configuration in `release-please-config.json` control release types
- [ ] Verify repository_dispatch integration between release.yml and publish.yml works correctly

**5.4 Do NOT Modify `release.yml`**

**File**: `.github/workflows/release.yml`

**Purpose**: Moonrepo's own release workflow

**Action**: NONE - This file is for moonrepo toolchain releases only

**Verification Checklist**:
- [ ] Confirm `release.yml` exists and is NOT modified
- [ ] Verify all changes are in `publish.yml` (or `release-please.yml`)
- [ ] Double-check PR diff doesn't include `release.yml` changes

### 6. Documentation

- [ ] Document workflow in `docs/development.md`:
  - How moonrepo detects affected projects
  - How to trigger releases (Conventional Commits)
  - How to test CI locally
  - New mise task structure and usage
  - How to pass package arguments to publish tasks
- [ ] Add CONTRIBUTING.md section on commit message format
- [ ] Update README with badge links to CI status
- [ ] Document mise task reorganization:
  - `.mise/tasks/node/publish` - Node/Bun packages
  - `.mise/tasks/go/publish` - Go binaries (future)
  - Examples: `mise run node:publish @zenobi-us/pkg1`

### 7. Testing

- [ ] Create test PR with changes to single package
- [ ] Verify only affected tests run
- [ ] Verify release-please creates correct PR
- [ ] Test actual release process on a test tag/version
- [ ] Test mise task invocation:
  - `mise run node:publish --dry-run`
  - `mise run node:publish @zenobi-us/opennotes`
  - Verify variadic arguments work correctly

## Future Mise Tasks

### Go Package Publishing

When Go packages need automated publishing, create `.mise/tasks/go/publish` as a **bash script**:

```bash
#!/usr/bin/env bash
# .mise/tasks/go/publish
# Description: Build and publish Go binaries
# Usage: mise run go:publish [service-name] [another-service]

# Accept variadic package names as arguments
PACKAGES=("$@")

if [ ${#PACKAGES[@]} -eq 0 ]; then
  echo "Building all Go services..."
  goreleaser release --clean
else
  echo "Building specific services: ${PACKAGES[*]}"
  for pkg in "${PACKAGES[@]}"; do
    echo "Building $pkg..."
    goreleaser release --clean --single-target --id="$pkg"
  done
fi
```

**Setup**:
```bash
chmod +x .mise/tasks/go/publish
```

**Benefits**:
- Consistent interface with `node:publish`
- Selective building based on affected packages
- Integration with GoReleaser for cross-compilation
- Variadic arguments for multiple services
- Simple bash syntax (no templating needed)

**Integration with CI**:
```yaml
- name: Publish affected Go services
  if: contains(steps.affected.outputs.types, 'go')
  run: mise run go:publish ${{ steps.affected.outputs.go_packages }}
```

### Other Runtime Tasks

As the monorepo grows, additional task namespaces can be created as **bash scripts**:

- `.mise/tasks/rust/publish` - Rust crate publishing
- `.mise/tasks/python/publish` - Python package publishing  
- `.mise/tasks/docker/publish` - Container image publishing

Each namespace follows the same pattern:
1. **Bash script format** (executable, no TOML wrapper)
2. **Variadic argument support** using `$@` for selective operations
3. **Integration with moonrepo** affected detection
4. **Consistent naming** and interface across runtimes

**Example Template**:
```bash
#!/usr/bin/env bash
# .mise/tasks/{runtime}/publish

PACKAGES=("$@")

if [ ${#PACKAGES[@]} -eq 0 ]; then
  echo "Publishing all {runtime} packages..."
  # Publish all logic here
else
  echo "Publishing specific packages: ${PACKAGES[*]}"
  for pkg in "${PACKAGES[@]}"; do
    echo "Publishing $pkg..."
    # Publish specific package logic here
  done
fi
```

### 8. Implementation Checklist

**Complete Workflow Implementation Checklist**:

#### Scripts and Tooling
- [ ] Create `.github/scripts/` directory
- [ ] Create `.github/scripts/get-affected-pkgs.sh` script
- [ ] Add script header with usage documentation
- [ ] Implement moonrepo query logic (`moon query projects --affected --json`)
- [ ] Implement jq filtering by package type (node, go, etc.)
- [ ] Add DRY_RUN mode for debugging
- [ ] Make script executable (`chmod +x .github/scripts/get-affected-pkgs.sh`)
- [ ] Test script locally with `DRY_RUN=true`
- [ ] Test script with different base branches
- [ ] Test script with different package types
- [ ] Verify script output format (space-separated package names)

#### Workflow Updates
- [ ] Update `.github/workflows/publish.yml` to call get-affected-pkgs.sh
- [ ] Replace inline bash logic with script invocation
- [ ] Add quality gate job using moonrepo CI
- [ ] Add conditional execution based on script output
- [ ] Test workflow with sample PR (dry-run)
- [ ] Verify GitHub Actions can execute the script
- [ ] Confirm script outputs are correctly captured in workflow

#### Documentation
- [ ] Document script usage in comments
- [ ] Update CONTRIBUTING.md with testing instructions
- [ ] Add examples of local testing commands
- [ ] Document DRY_RUN mode for debugging
- [ ] Explain integration with mise tasks
- [ ] Add troubleshooting section for common issues

#### Validation
- [ ] Verify script works with no affected packages (empty output)
- [ ] Verify script works with single affected package
- [ ] Verify script works with multiple affected packages
- [ ] Verify script filters by package type correctly
- [ ] Confirm script output integrates with `mise run node:publish`
- [ ] Test complete pipeline: commit → detect → publish

## Testing Scripts Locally

**Purpose**: Validate CI/CD scripts before pushing to GitHub to avoid trial-and-error debugging in Actions

### Testing `get-affected-pkgs.sh`

**Location**: `.github/scripts/get-affected-pkgs.sh`

**Basic Testing**:
```bash
# 1. Test with dry-run mode (verbose debug output)
DRY_RUN=true bash .github/scripts/get-affected-pkgs.sh main node

# Expected output (stderr):
# DEBUG: Base branch: main
# DEBUG: Package type: node
# DEBUG: Affected projects JSON: [...]
# DEBUG: Filtered packages: @zenobi-us/pkg1 @zenobi-us/pkg2

# 2. Test actual output (what CI receives)
bash .github/scripts/get-affected-pkgs.sh main node

# Expected output (stdout): @zenobi-us/pkg1 @zenobi-us/pkg2
```

**Testing Different Scenarios**:

```bash
# Test with no changes (should output empty string)
git checkout main
bash .github/scripts/get-affected-pkgs.sh main node
# Expected: "" (empty)

# Test with changes to a Node package
git checkout -b test-branch
# Make changes to a node package
touch pkgs/pi-opennotes/test-file.ts
git add .
git commit -m "test: trigger affected detection"
bash .github/scripts/get-affected-pkgs.sh main node
# Expected: "pkgs/pi-opennotes" or "@zenobi-us/pi-opennotes"

# Test with changes to a Go package
touch services/go-api/test.go
git add .
git commit -m "test: trigger go affected"
bash .github/scripts/get-affected-pkgs.sh main go
# Expected: "services/go-api" or "go-api"

# Test with multiple affected packages
# Make changes to multiple packages, test output

# Clean up test branch
git checkout main
git branch -D test-branch
```

**Testing jq Filter Logic**:

```bash
# Test the jq filter independently
AFFECTED=$(moon query projects --affected --json --base main)

# Test Node package filter
echo "$AFFECTED" | jq -r '.[] | select(.language == "node") | .id'

# Test Go package filter  
echo "$AFFECTED" | jq -r '.[] | select(.language == "go") | .id'

# Test with tags instead of language (if your project uses tags)
echo "$AFFECTED" | jq -r '.[] | select(.tags[] | contains("node")) | .id'
```

**Integration Testing with Mise**:

```bash
# Test the full pipeline locally
PACKAGES=$(bash .github/scripts/get-affected-pkgs.sh main node)
echo "Affected packages: $PACKAGES"

# Test mise task invocation
mise run node:publish --dry-run $PACKAGES

# Or test with explicit package names
mise run node:publish @zenobi-us/pkg1 @zenobi-us/pkg2
```

**Common Issues and Debugging**:

| Issue | Symptom | Solution |
|-------|---------|----------|
| Empty output | Script returns nothing | Set `DRY_RUN=true` to see debug output |
| Wrong packages detected | Script returns unexpected packages | Check jq filter logic matches moonrepo structure |
| Script errors | Bash errors or jq errors | Check `set -euo pipefail`, verify moonrepo installed |
| Trailing whitespace | Extra spaces in output | Script uses `xargs` to trim, verify it's working |

**Best Practices**:

1. **Always test locally first**: Don't push untested scripts to CI
2. **Use DRY_RUN mode**: Enables verbose debugging without modifying script
3. **Test edge cases**: No changes, single package, multiple packages, all packages
4. **Verify output format**: Ensure output matches what mise tasks expect
5. **Test with different branches**: Verify `--base` parameter works correctly
6. **Validate jq filters**: Test jq logic independently before embedding in script
7. **Check script permissions**: Ensure script is executable (`chmod +x`)

## Expected Outcome

After completion:

### Workflow Architecture

1. **PR Validation** (`.github/workflows/pr.yml`)
   - ✅ PRs automatically run tests for affected packages only
   - ✅ Moonrepo detects affected packages based on git diff
   - ✅ Dependent packages are tested to prevent breaking changes
   - ✅ Tests must pass before merge allowed

2. **Package Publishing** (`.github/workflows/publish.yml`)
   - ✅ Quality gate runs first (tests affected packages)
   - ✅ Detects affected packages using `moon query projects --affected`
   - ✅ Filters packages by type (node vs go)
   - ✅ Calls `mise run node:publish` with variadic package arguments
   - ✅ Only publishes packages that have changes and pass tests
   - ✅ GoReleaser integration for Go binaries (triggered by tags)

3. **Version Management** (release-please)
   - ✅ Automatic version bumps using Conventional Commits
   - ✅ Independent versions per package (no linked versions)
   - ✅ Automatic changelog generation per package
   - ✅ Creates release PRs that, when merged, trigger publishing

4. **Mise Task Structure**
   - ✅ Namespaced tasks: `.mise/tasks/node/publish`, `.mise/tasks/go/publish`
   - ✅ Variadic arguments: `mise run node:publish pkg1 pkg2`
   - ✅ Integration with moonrepo affected detection
   - ✅ Scalable for future runtimes (rust, python, etc.)

### Concrete Benefits

1. **Safety**: Cannot release if dependent packages have test failures
2. **Efficiency**: Only affected projects are tested/built (reduced CI time)
3. **Independence**: Each package versions independently when it has changes
4. **Clarity**: Clear feedback on what's being tested and why
5. **Automation**: Release process driven by commit messages (Conventional Commits)
6. **Consistency**: Unified publish interface across runtimes via mise tasks

### Developer Experience

**Before this work**:
- Manual version bumps
- Unclear what needs testing
- Risk of breaking dependent packages
- Manual changelog maintenance
- Single publish task for all package types

**After this work**:
- Automatic version bumps via commit messages
- Clear indication of affected packages
- Impossible to release with broken dependencies
- Automatic changelogs
- Runtime-specific publish tasks with clear interface

## Actual Outcome

_To be filled after implementation_

## Lessons Learned

_To be filled after implementation_

## GoReleaser Integration for Go Packages

### Overview

GoReleaser provides native monorepo support that can work alongside our release-please + moonrepo strategy. This is particularly relevant for our Go packages, enabling automated binary builds, cross-compilation, and distribution.

### Key GoReleaser Monorepo Features

1. **Path-based filtering**: Can build/release only changed Go modules
2. **Multiple .goreleaser.yml files**: Support for per-package configuration
3. **Integration with tag patterns**: Works with tags like `service-name/v1.2.3`
4. **Build matrix support**: Can handle multiple Go modules in one release workflow

### Integration Points with release-please

**Potential Workflow**:
1. release-please creates version bump PR with Conventional Commits
2. Merging PR triggers release-please to create tags (e.g., `go-api-v1.2.3`)
3. Tag creation triggers GoReleaser workflow
4. GoReleaser builds binaries, creates GitHub release, attaches artifacts

**Benefits**:
- release-please handles version management and changelogs
- GoReleaser handles binary compilation and distribution
- Clean separation of concerns
- Both tools play to their strengths

### Research Questions

These questions need to be answered during implementation:

1. **Tag Pattern Compatibility**: How does GoReleaser fit with release-please's tag creation?
   - Does release-please create tags in a format GoReleaser expects?
   - Do we need custom tag patterns for monorepo packages?

2. **Tool Division of Labor**: Should we use GoReleaser for Go packages and release-please for version management?
   - Or does release-please handle everything (version + release creation)?
   - What's the cleanest workflow separation?

3. **Tag Consumption**: Can GoReleaser consume the tags that release-please creates?
   - Do tag patterns need alignment (e.g., `go-api/v1.2.3` vs `go-api-v1.2.3`)?
   - Does GoReleaser need configuration to match release-please's tag format?

4. **Workflow Separation**: Do we need separate release workflows for Go vs Bun/Node packages?
   - Go packages → release-please (version) + GoReleaser (build/publish)
   - Node packages → release-please (version + npm publish)
   - Or unified workflow with conditional steps?

5. **Multiple .goreleaser.yml Files**: Should each Go package have its own GoReleaser config?
   - Pro: Package-specific build configurations
   - Con: Configuration duplication
   - Alternative: Single config with monorepo filtering

### Implementation Considerations

**Recommended Approach** (to be validated):

```yaml
# Workflow triggered by tag creation
on:
  push:
    tags:
      - 'go-api-v*'
      - 'other-go-package-v*'

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      
      - uses: actions/setup-go@v5
      
      - uses: goreleaser/goreleaser-action@v5
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
          workdir: ./services/go-api  # Path-based filtering
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**GoReleaser Configuration** (`.goreleaser.yml` in package directory):

```yaml
# services/go-api/.goreleaser.yml
project_name: opennotes

builds:
  - main: ./main.go
    binary: opennotes
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

release:
  github:
    owner: zenobi-us
    name: opennotes
  name_template: "go-api v{{ .Version }}"
```

### Decision Points

Before implementing, decide:

1. **Tag Format**: Align release-please and GoReleaser on tag patterns
   - Recommended: `go-api-v1.2.3` (works with both tools)

2. **Workflow Trigger**: When should GoReleaser run?
   - Option A: On tag creation (after release-please creates tag)
   - Option B: After release-please creates release (duplicate effort?)

3. **Binary Distribution**: Where should Go binaries be published?
   - GitHub Releases (GoReleaser default)
   - Package registries (Homebrew, apt, etc.)
   - Docker images

4. **Monorepo Strategy**: Single or multiple GoReleaser configs?
   - Recommendation: Per-package configs for flexibility

### Action Items for Implementation

- [ ] Review GoReleaser monorepo documentation thoroughly
- [ ] Align tag patterns between release-please and GoReleaser
- [ ] Create `.goreleaser.yml` for `services/go-api`
- [ ] Add GoReleaser workflow triggered by tag creation
- [ ] Test workflow with test tag
- [ ] Document GoReleaser + release-please integration

## References

- [moonrepo Documentation](https://moonrepo.dev/)
- [release-please Documentation](https://github.com/googleapis/release-please)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Google release-please Action](https://github.com/google-github-actions/release-please-action)
- [GoReleaser Monorepo Support](https://goreleaser.com/customization/monorepo/)
- [Mise Task Arguments - Variadic Arguments](https://mise.jdx.dev/tasks/task-arguments.html#variadic-arguments)

## Notes

### Workflow File Clarification

**CRITICAL**: This task modifies `.github/workflows/publish.yml`, NOT `release.yml`.

- `publish.yml` - Custom CI/CD workflow for our package publishing
- `release.yml` - Moonrepo's own release process (do not modify)

### Package Structure Assumptions

This task assumes the following package structure:
- `services/go-api/` - Go CLI application (main OpenNotes binary)
- `apps/web-app/` - Bun/TypeScript web application (example dependent)
- `pkgs/pi-opennotes/` - Bun/TypeScript pi extension package

Adjust configurations if actual structure differs.

### Mise Task Reorganization Rationale

Moving from `.mise/tasks/publish` to `.mise/tasks/node/publish`:

**Benefits**:
1. **Namespace Clarity**: Tasks grouped by runtime/language
2. **Scalability**: Easy to add `go/publish`, `rust/publish`, etc.
3. **Selective Invocation**: `mise run node:publish` vs `mise run go:publish`
4. **Variadic Arguments**: Pass specific packages to publish using bash `$@`
5. **Integration**: Works seamlessly with moonrepo affected detection
6. **Bash Script Format**: Simple, direct bash scripts (no TOML wrapper needed)

**File Format**:
- **Keep as bash scripts**: `.mise/tasks/node/publish` (executable bash file, no extension)
- **NOT TOML files**: Mise executes bash scripts directly as tasks
- **Variadic args in bash**: Use `$@` to accept multiple package arguments

**Migration Path**:
```bash
# Old way
mise run publish

# New way (bash script with variadic args)
mise run node:publish
mise run node:publish @zenobi-us/pkg1 @zenobi-us/pkg2
```

### Conventional Commits Requirement

For release-please to work correctly, the project MUST use Conventional Commits:
- `feat(go-api): add new flag` → minor version bump for go-api
- `fix(web-app): correct typo` → patch version bump for web-app
- `feat!: breaking change` → major version bump

### Go Version Management

Go doesn't have a standard version file like `package.json`. Options:
1. Use `version.go` with a const (recommended for simplicity)
2. Use build flags: `-ldflags "-X main.Version=1.0.0"`
3. Use go.mod version tag (less common)

The task uses option 1 for simplicity and clarity.

### Deployment Targets

Consider where releases should be published:
- **Go Binary**: GitHub Releases with compiled binaries (linux, macOS, windows)
- **Node/Bun Packages**: npm registry as scoped packages (@zenobi-us/...)
- **Docker Images**: Optional, for containerized deployments

Each publish target may need separate workflow jobs.
