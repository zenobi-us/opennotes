# Phase 4: Polish

## Overview
Final polish including error handling, documentation, testing, and build configuration.

## Status: PARTIAL (6/9 tasks complete)

## Tasks

### 4.1 Error Handling
- [x] Review and standardize error messages across all commands
- [x] Add context to errors using `fmt.Errorf("context: %w", err)`
- [x] Ensure all errors are logged appropriately
- [x] User-friendly error messages for common failures

### 4.2 Input Validation
- [x] Create `internal/core/schema.go` with validation helpers
- [x] `ValidateNotebookName()` - checks name format and length
- [x] `ValidatePath()` - checks for invalid characters
- [x] `ValidateNoteName()` - checks note filename, prevents traversal
- [x] `Validator` type for collecting multiple errors
- [x] `ValidationErrors` type with PrettyPrint()

### 4.3 String Utilities
- [x] Create `internal/core/strings.go`
- [x] Implement `Slugify()` - URL-friendly slug generation
- [x] Implement `Dedent()` - removes common leading whitespace
- [x] Implement `ObjectToFrontmatter()` - converts map to YAML

### 4.4 Help Text & Documentation
- [ ] Write comprehensive help text for all commands
- [ ] Add examples to command descriptions
- [ ] Document environment variables in help
- [ ] Match TypeScript CLI help format

### 4.5 Init Command
- [x] Create `cmd/init.go`
- [x] Initialize config file at `~/.config/opennotes/config.json`
- [x] Display success message with config path

### 4.6 Integration Tests
- [x] Set up test framework with testify
- [x] Create `internal/core/strings_test.go`
  - Test Slugify with various inputs
  - Test Dedent with various indentation
  - Test ObjectToFrontmatter
- [x] Create `internal/core/schema_test.go`
  - Test ValidateNotebookName
  - Test ValidateNoteName
  - Test ValidationErrors and Validator
- [ ] Create `internal/services/config_test.go`
  - Test default loading
  - Test env var override
  - Test file persistence
- [ ] Create `internal/services/notebook_test.go`
  - Test notebook creation
  - Test notebook discovery
  - Test context matching
- [ ] Create `internal/services/note_test.go`
  - Test note search
  - Test note count
- [ ] Create e2e tests for CLI commands

### 4.7 Build Configuration
- [x] Create mise tasks for Go build
- [x] Add `go-build` task: `go build -o dist/opennotes .`
- [x] Add `go-test` task: `go test ./...`
- [x] Add `go-lint` task with go vet/gofmt fallback
- [ ] Update CI/CD for Go builds

### 4.8 Configuration Compatibility
- [ ] Verify Go reads TypeScript config files correctly
- [ ] Verify Go writes config in same format as TypeScript
- [ ] Test notebook config `.opennotes.json` compatibility

### 4.9 Feature Parity Verification
- [ ] Compare all commands between Go and TypeScript versions
- [ ] Verify identical CLI interface (flags, args, output)
- [ ] Document any intentional differences

## Dependencies
- Phase 1-3 complete

## Acceptance Criteria
- [x] Core tests pass
- [ ] Service tests pass
- [ ] Help text is clear and complete
- [x] Error messages are user-friendly
- [ ] Config files are compatible between Go and TypeScript
- [x] Build produces working binary
- [ ] CI/CD pipeline builds Go version
