import { type } from 'arktype';
import type { ConfigService } from './ConfigService.ts';
import type { DbService } from './Db.ts';
import { VARCHAR } from '@duckdb/node-api';
import { Logger } from './LoggerService.ts';
import { join } from 'node:path';

const _NoteSchema = type({
  metadata: type({ '[string]': 'string | number | boolean' }),
  content: 'string',
});

const _NotebookMetadataSchema = type({
  '[string]': 'string | number | boolean',
});
export type NotebookMetadata = typeof _NotebookMetadataSchema.infer;
export type Note = typeof _NoteSchema.infer;

const Log = Logger.child({ namespace: 'NoteService' });

export type NoteService = ReturnType<typeof createNoteService>;

export function createNoteService(options: {
  dbService: DbService;
  notebookPath?: string;
  configService: ConfigService;
}) {
  async function query(query: string) {
    const database = await options.dbService.getDb();

    const result = await database.run(query);

    if (!result) {
      return [];
    }

    return result;
  }

  /**
   * Reads a markdown note by ID from the notebook path.
   * @param noteId - The note identifier (filename without extension)
   * @returns The markdown content as a string, or null if not found
   */
  async function readNote(filepath: string) {
    const db = await options.dbService.getDb();

    try {
      const prepared = await db.prepare(`SELECT content, metadata FROM read_markdown('$filepath')`);
      prepared.bind({ filepath });
      const result = await prepared.run();

      const rows = await result.getRowObjectsJson();

      if (rows?.length === 0) {
        return null;
      }
      if (!rows[0]) {
        return null;
      }

      return rows[0];
    } catch {
      return null;
    }
  }

  /**
   * Searches notes using a user-provided DuckDB SQL query.
   * The query should reference the notebook path and can use markdown functions.
   * @param query - Raw DuckDB SQL query
   * @returns Array of note IDs (filenames without extension) matching the query
   */
  async function searchNotes(args?: { query?: string }) {
    if (!options.notebookPath) {
      throw new Error('No notebook selected');
    }

    Log.debug('searchNotes: query=%s notebookPath=%s', args?.query, options.notebookPath);

    const db = await options.dbService.getDb();
    const prepared = await db.prepare(`
      SELECT * FROM read_markdown($filepath)
    `);
    prepared.bind(
      {
        filepath: join(options.notebookPath, '**', '*.md'),
      },
      {
        filepath: VARCHAR,
      }
    );

    const result = await prepared.run();

    const rows = await result.getRowObjectsJson();

    return rows || [];
  }

  /**
   * Create a new notebook template
   */
  function _createTemplate(): void {
    // Not yet implemented
  }

  /**
   * TODO: implement note creation
   */
  async function createNote(): Promise<void> {
    // Not implemented
  }

  /**
   * TODO: implement note removal
   */
  async function removeNote(): Promise<void> {
    // Not implemented
  }

  /**
   * TODO: implement note editing
   */
  async function editNote(): Promise<void> {
    // Not implemented
  }

  /**
   * TODO: implement note counting
   */
  const CountResultSchema = type({
    count: 'string.integer.parse',
  }).array();

  async function count(): Promise<number> {
    // Not implemented
    const db = await options.dbService.getDb();
    const notebookPath = options.notebookPath || '';
    const glob = join(notebookPath, '**', '*.md');
    const prepared = await db.prepare(`SELECT COUNT(*) as count FROM read_markdown($pattern)`);
    Log.debug('count: notebookPath=%s glob=%s', notebookPath, glob);
    prepared.bind(
      {
        pattern: glob,
      },
      {
        pattern: VARCHAR,
      }
    );

    const result = await prepared.run();
    const rows = await result.getRowObjectsJson();

    const parsed = CountResultSchema(rows);
    if (parsed instanceof type.errors) {
      Log.error('count: failed to parse count result: %o', parsed);
      return 0;
    }
    Log.debug('count: rows=%o', rows);
    const count = parsed[0]?.count || 0;

    return count;
  }

  return {
    createNote,
    readNote,
    removeNote,
    editNote,
    searchNotes,
    query,
    count,
  };
}
