#!/usr/bin/env bats
# Jot Workflow Enforcement E2E Tests
# Tests CLI behavior when workflow rules block or allow operations
# Exit code 4 = ExitCodeWorkflowBlocked

setup() {
    export TEST_DIR="$(mktemp -d)"
    export HOME_BACKUP="$HOME"
    export HOME="$TEST_DIR"
    export JOT_CONFIG="$TEST_DIR/.config/jot/config.json"
    
    # Ensure binary exists
    if [[ ! -f "../../dist/jot" ]]; then
        cd ../.. && mise run build && cd tests/e2e
    fi
    
    export PATH="$(pwd)/../../dist:$PATH"
    
    # Initialize jot
    jot init >/dev/null 2>&1
}

teardown() {
    rm -rf "$TEST_DIR"
    export HOME="$HOME_BACKUP"
}

# Helper: Create a notebook with workflow enforcement
# Uses a multi-step workflow: todo -> in-progress -> done
# This ensures we can test blocked transitions properly
create_workflow_notebook() {
    local notebook_dir="$TEST_DIR/workflow-test"
    mkdir -p "$notebook_dir/.notes"
    
    # Create .jot.json with groups and workflows
    cat > "$notebook_dir/.jot.json" <<'EOF'
{
  "config_version": 1,
  "root": ".",
  "name": "Workflow Test",
  "groups": [
    {
      "name": "tasks",
      "globs": ["tasks/*.md"],
      "workflow_id": "task_workflow"
    },
    {
      "name": "docs",
      "globs": ["docs/*.md"]
    }
  ],
  "workflows": {
    "task_workflow": {
      "description": "Task workflow with state transitions",
      "initial_state": "todo",
      "mode": "enforce",
      "field": "status",
      "states": {
        "todo": {
          "schema": {"type": "object", "required": ["title"]},
          "transitions": ["in-progress", "cancelled"]
        },
        "in-progress": {
          "schema": {"type": "object", "required": ["title", "assignee"]},
          "transitions": ["done", "blocked", "todo"]
        },
        "blocked": {
          "schema": {"type": "object", "required": ["title", "blocked_reason"]},
          "transitions": ["in-progress", "cancelled"]
        },
        "done": {
          "schema": {"type": "object", "required": ["title", "completed_at"]},
          "transitions": []
        },
        "cancelled": {
          "schema": {"type": "object", "required": ["title"]},
          "transitions": []
        }
      }
    }
  }
}
EOF
    
    # Register the notebook
    jot notebook register "$notebook_dir" >/dev/null 2>&1
    
    echo "$notebook_dir"
}

# Helper: Create a note file with frontmatter
create_note_with_frontmatter() {
    local dir="$1"
    local filename="$2"
    local frontmatter="$3"
    local body="${4:-# Test Note}"
    
    mkdir -p "$(dirname "$dir/$filename")"
    cat > "$dir/$filename" <<EOF
---
$frontmatter
---

$body
EOF
}

# =============================================================================
# EXIT CODE 4: Workflow Blocked Tests
# =============================================================================

@test "Exit code 4: workflow blocks note creation with invalid initial state" {
    notebook_dir="$(create_workflow_notebook)"
    
    # Try to create a task with invalid initial status (done instead of todo)
    # todo can only transition to in-progress or cancelled, not done
    run jot --notebook "$notebook_dir" notes add "Bad Task" "tasks/bad-task.md" \
        --data status=done --data title="Bad Task"
    
    # Should fail with exit code 4 (workflow blocked)
    [[ "$status" -eq 4 ]]
    [[ "$output" =~ "workflow" ]]
    [[ "$output" =~ "blocked" ]]
}

@test "Exit code 4: workflow blocks invalid state transition on update" {
    notebook_dir="$(create_workflow_notebook)"
    
    # Create a valid task in 'todo' state
    create_note_with_frontmatter "$notebook_dir" "tasks/my-task.md" \
        "title: My Task
status: todo"
    
    # Try to transition directly from todo → done (invalid, must go through in-progress)
    # Note: BATS run doesn't work with piped input - wrap in bash -c
    run bash -c "echo -e '---\ntitle: My Task\nstatus: done\ncompleted_at: 2026-03-01\n---\n\n# My Task' | jot --notebook '$notebook_dir' notes update 'tasks/my-task.md'"
    
    # Should fail with exit code 4
    [[ "$status" -eq 4 ]]
    [[ "$output" =~ "workflow" ]] || [[ "$output" =~ "blocked" ]]
}

@test "Exit code 4: workflow blocks transition missing required metadata" {
    notebook_dir="$(create_workflow_notebook)"
    
    # Create a valid task in 'todo' state
    create_note_with_frontmatter "$notebook_dir" "tasks/missing-meta.md" \
        "title: Missing Meta Task
status: todo"
    
    # Try to transition todo → in-progress without required 'assignee' field
    run bash -c "echo -e '---\ntitle: Missing Meta Task\nstatus: in-progress\n---\n\n# Missing Meta Task' | jot --notebook '$notebook_dir' notes update 'tasks/missing-meta.md'"
    
    # Should fail with exit code 4 (missing required field per state schema)
    [[ "$status" -eq 4 ]]
}

# =============================================================================
# SUCCESSFUL WORKFLOW OPERATIONS
# =============================================================================

@test "Workflow allows valid initial state creation" {
    notebook_dir="$(create_workflow_notebook)"
    
    # Create a task with valid initial status (todo)
    run jot --notebook "$notebook_dir" notes add "Good Task" "tasks/good-task.md" \
        --data status=todo --data title="Good Task"
    
    # Should succeed
    [[ "$status" -eq 0 ]]
    [[ -f "$notebook_dir/tasks/good-task.md" ]]
}

@test "Workflow allows valid state transition" {
    notebook_dir="$(create_workflow_notebook)"
    
    # Create a valid task in 'todo' state
    create_note_with_frontmatter "$notebook_dir" "tasks/valid-transition.md" \
        "title: Valid Transition Task
status: todo"
    
    # Transition todo → in-progress with required assignee
    run bash -c "echo -e '---\ntitle: Valid Transition Task\nstatus: in-progress\nassignee: alice\n---\n\n# Valid Transition Task' | jot --notebook '$notebook_dir' notes update 'tasks/valid-transition.md'"
    
    # Should succeed
    [[ "$status" -eq 0 ]]
    
    # Verify the file was updated
    grep -q "status: in-progress" "$notebook_dir/tasks/valid-transition.md"
}

@test "Workflow allows complete lifecycle: todo → in-progress → done" {
    notebook_dir="$(create_workflow_notebook)"
    
    # Create task in todo
    run jot --notebook "$notebook_dir" notes add "Lifecycle Task" "tasks/lifecycle.md" \
        --data status=todo --data title="Lifecycle Task"
    [[ "$status" -eq 0 ]]
    
    # Transition to in-progress
    run bash -c "echo -e '---\ntitle: Lifecycle Task\nstatus: in-progress\nassignee: bob\n---\n\n# Lifecycle Task' | jot --notebook '$notebook_dir' notes update 'tasks/lifecycle.md'"
    [[ "$status" -eq 0 ]]
    
    # Transition to done
    run bash -c "echo -e '---\ntitle: Lifecycle Task\nstatus: done\nassignee: bob\ncompleted_at: 2026-03-01\n---\n\n# Lifecycle Task' | jot --notebook '$notebook_dir' notes update 'tasks/lifecycle.md'"
    [[ "$status" -eq 0 ]]
    
    # Verify final state
    grep -q "status: done" "$notebook_dir/tasks/lifecycle.md"
}

# =============================================================================
# NOTES WITHOUT WORKFLOW BINDING
# =============================================================================

@test "Notes outside workflow groups are allowed without restrictions" {
    notebook_dir="$(create_workflow_notebook)"
    
    # Create a note in docs/ which has no workflow binding
    run jot --notebook "$notebook_dir" notes add "Free Doc" "docs/free-doc.md" \
        --data status=anything --data random_field="whatever"
    
    # Should succeed - no workflow enforcement
    [[ "$status" -eq 0 ]]
    [[ -f "$notebook_dir/docs/free-doc.md" ]]
}

@test "Notes not matching any group are allowed without restrictions" {
    notebook_dir="$(create_workflow_notebook)"
    
    # Create a note at root level (not matching any group glob)
    run jot --notebook "$notebook_dir" notes add "Root Note" "random-note.md" \
        --data status=invalid_state
    
    # Should succeed - not in any group
    [[ "$status" -eq 0 ]]
    [[ -f "$notebook_dir/random-note.md" ]]
}

# =============================================================================
# ERROR MESSAGE QUALITY
# =============================================================================

@test "Error message includes workflow diagnostic code" {
    notebook_dir="$(create_workflow_notebook)"
    
    # Try invalid transition (todo → done directly)
    run jot --notebook "$notebook_dir" notes add "Diagnostic Test" "tasks/diagnostic-test.md" \
        --data status=done --data title="Diagnostic Test"
    
    [[ "$status" -eq 4 ]]
    # Error should contain diagnostic info
    [[ "$output" =~ "invalid_transition" ]] || [[ "$output" =~ "not allowed" ]]
}

@test "Error message identifies blocking workflow" {
    notebook_dir="$(create_workflow_notebook)"
    
    run jot --notebook "$notebook_dir" notes add "Bad Initial" "tasks/bad-initial.md" \
        --data status=done --data title="Bad Initial"
    
    [[ "$status" -eq 4 ]]
    # Should mention which workflow blocked it
    [[ "$output" =~ "task_workflow" ]]
}

# =============================================================================
# TERMINAL STATES
# =============================================================================

@test "Exit code 4: cannot transition from terminal state (done)" {
    notebook_dir="$(create_workflow_notebook)"
    
    # Create a completed task
    create_note_with_frontmatter "$notebook_dir" "tasks/completed-task.md" \
        "title: Completed Task
status: done
completed_at: 2026-02-28"
    
    # Try to change status back to in-progress
    run bash -c "echo -e '---\ntitle: Completed Task\nstatus: in-progress\nassignee: alice\n---\n\n# Completed Task' | jot --notebook '$notebook_dir' notes update 'tasks/completed-task.md'"
    
    # Should fail - 'done' is a terminal state with no transitions
    [[ "$status" -eq 4 ]]
}

@test "Exit code 4: cannot transition from terminal state (cancelled)" {
    notebook_dir="$(create_workflow_notebook)"
    
    # Create a cancelled task
    create_note_with_frontmatter "$notebook_dir" "tasks/cancelled-task.md" \
        "title: Cancelled Task
status: cancelled"
    
    # Try to revive it
    run bash -c "echo -e '---\ntitle: Cancelled Task\nstatus: todo\n---\n\n# Cancelled Task' | jot --notebook '$notebook_dir' notes update 'tasks/cancelled-task.md'"
    
    # Should fail - 'cancelled' is a terminal state
    [[ "$status" -eq 4 ]]
}

# =============================================================================
# BLOCKED STATE WORKFLOW
# =============================================================================

@test "Workflow allows blocked state with required fields" {
    notebook_dir="$(create_workflow_notebook)"
    
    # Create task in in-progress
    create_note_with_frontmatter "$notebook_dir" "tasks/blockable.md" \
        "title: Blockable Task
status: in-progress
assignee: charlie"
    
    # Transition to blocked with required blocked_reason
    run bash -c "echo -e '---\ntitle: Blockable Task\nstatus: blocked\nblocked_reason: Waiting for API access\n---\n\n# Blockable Task' | jot --notebook '$notebook_dir' notes update 'tasks/blockable.md'"
    
    [[ "$status" -eq 0 ]]
    grep -q "status: blocked" "$notebook_dir/tasks/blockable.md"
}

@test "Exit code 4: blocked state without required reason" {
    notebook_dir="$(create_workflow_notebook)"
    
    # Create task in in-progress
    create_note_with_frontmatter "$notebook_dir" "tasks/no-reason.md" \
        "title: No Reason Task
status: in-progress
assignee: dave"
    
    # Try to transition to blocked without blocked_reason
    run bash -c "echo -e '---\ntitle: No Reason Task\nstatus: blocked\n---\n\n# No Reason Task' | jot --notebook '$notebook_dir' notes update 'tasks/no-reason.md'"
    
    # Should fail - missing required blocked_reason
    [[ "$status" -eq 4 ]]
}

# =============================================================================
# EDGE CASES
# =============================================================================

@test "Creating note with no status field uses workflow initial_state" {
    notebook_dir="$(create_workflow_notebook)"
    
    # Create without specifying status - should default to initial_state (todo)
    run jot --notebook "$notebook_dir" notes add "No Status" "tasks/no-status.md" \
        --data title="No Status"
    
    # Should succeed (self-transition todo -> todo is allowed)
    [[ "$status" -eq 0 ]]
}

@test "Self-transition (same state) is always allowed" {
    notebook_dir="$(create_workflow_notebook)"
    
    # Create task in todo
    create_note_with_frontmatter "$notebook_dir" "tasks/self-trans.md" \
        "title: Self Trans Task
status: todo"
    
    # Update without changing status (todo -> todo)
    run bash -c "echo -e '---\ntitle: Updated Self Trans Task\nstatus: todo\n---\n\n# Updated' | jot --notebook '$notebook_dir' notes update 'tasks/self-trans.md'"
    
    # Should succeed
    [[ "$status" -eq 0 ]]
}
