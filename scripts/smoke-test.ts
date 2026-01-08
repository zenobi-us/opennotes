#!/usr/bin/env bun
/**
 * Smoke test for opennotes WASM migration
 * Tests basic functionality after build
 */

import { existsSync, mkdirSync, rmSync, writeFileSync } from 'fs';
import { join } from 'path';
import { tmpdir } from 'os';

const OPENNOTES_BIN = join(process.cwd(), 'dist', 'opennotes');

async function runCommand(
  args: string[],
  options?: { cwd?: string }
): Promise<{ stdout: string; stderr: string; exitCode: number }> {
  const proc = Bun.spawn({
    cmd: [OPENNOTES_BIN, ...args],
    cwd: options?.cwd || process.cwd(),
    stdout: 'pipe',
    stderr: 'pipe',
  });

  const stdout = await new Response(proc.stdout).text();
  const stderr = await new Response(proc.stderr).text();
  const exitCode = await proc.exited;

  return { stdout, stderr, exitCode: exitCode || 0 };
}

async function main() {
  console.log('🧪 Running opennotes WASM smoke tests...\n');

  if (!existsSync(OPENNOTES_BIN)) {
    console.error(`❌ Binary not found at ${OPENNOTES_BIN}`);
    process.exit(1);
  }

  const tmpDir = join(tmpdir(), `opennotes-smoke-${Date.now()}`);
  const notebookDir = join(tmpDir, 'test-notebook');

  try {
    // Test 1: Create notebook structure
    console.log('📝 Test 1: Notebook Creation');
    mkdirSync(notebookDir, { recursive: true });

    // Create .opennotes.json at notebook root with contexts to enable discovery
    writeFileSync(
      join(notebookDir, '.opennotes.json'),
      JSON.stringify(
        {
          name: 'Test Notebook',
          description: 'Smoke test notebook',
          root: '.', // Notes are in current directory
          contexts: [notebookDir], // Add this directory as a context for discovery
        },
        null,
        2
      )
    );

    // Create markdown files in the notebook directory
    writeFileSync(join(notebookDir, 'note1.md'), '# Note 1\n\nFirst note');
    writeFileSync(join(notebookDir, 'note2.md'), '# Note 2\n\nSecond note');
    writeFileSync(join(notebookDir, 'note3.md'), '# Note 3\n\nThird note');

    console.log('   ✅ Notebook created with 3 markdown files\n');

    // Test 2: List notes from notebook directory (tests WASM DB query)
    console.log('📝 Test 2: Notes List (WASM DB Query)');
    const result = await runCommand(['notes', 'list'], {
      cwd: notebookDir,
    });

    if (result.exitCode !== 0) {
      console.error(`   ❌ Failed with exit code ${result.exitCode}`);
      if (result.stderr) console.error(`   stderr: ${result.stderr.substring(0, 300)}`);
      console.error(`   stdout: ${result.stdout.substring(0, 300)}`);
      process.exit(1);
    }

    // Check if the output contains the note files
    const noteCount = (result.stdout.match(/note\d\.md/g) || []).length;
    if (noteCount !== 3) {
      console.error(`   ❌ Expected 3 notes, found ${noteCount}`);
      console.error(`   Full output: ${result.stdout}`);
      process.exit(1);
    }

    console.log(`   ✅ Found and listed 3 notes via WASM DuckDB\n`);

    // Test 3: Test notebook info command
    console.log('📝 Test 3: Notebook Info');
    const infoResult = await runCommand(['notebook'], {
      cwd: notebookDir,
    });

    if (infoResult.exitCode !== 0) {
      console.error(`   ❌ Failed with exit code ${infoResult.exitCode}`);
      if (infoResult.stderr) console.error(`   stderr: ${infoResult.stderr.substring(0, 300)}`);
      process.exit(1);
    }

    if (!infoResult.stdout.includes('Test Notebook')) {
      console.error(`   ❌ Expected notebook name in output`);
      console.error(`   Output: ${infoResult.stdout}`);
      process.exit(1);
    }

    console.log('   ✅ Notebook info retrieval works\n');

    console.log('================');
    console.log('✅ All smoke tests passed!');
    console.log('   WASM DuckDB integration verified');
    console.log('================');
  } finally {
    rmSync(tmpDir, { recursive: true, force: true });
  }
}

main().catch((err) => {
  console.error('❌ Error:', err);
  process.exit(1);
});
