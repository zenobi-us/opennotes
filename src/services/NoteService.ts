import { type } from 'arktype';
import type { ConfigService } from './ConfigService.ts';
import type { DbService } from './Db';
import { Logger } from './LoggerService.ts';
import { join } from 'node:path';
import { TuiRender } from './Display.ts';
import { dedent } from '../core/strings.ts';
import { prettifyArktypeErrors } from '../core/schema.ts';

const NoteDbSchema = type({
  file_path: 'string',
  metadata: type({ '[string]': 'string | number | boolean' }),
  content: 'string',
});

const NotebookMetadataSchema = type({
  '[string]': 'string | number | boolean',
});

const NoteSchema = type({
  file: type({
    filepath: 'string',
    relative: 'string',
  }),
  content: 'string',
  metadata: NotebookMetadataSchema,
});

export type NotebookMetadata = typeof NotebookMetadataSchema.infer;
export type Note = typeof NoteSchema.infer;

const Log = Logger.child({ namespace: 'NoteService' });

export type NoteService = ReturnType<typeof createNoteService>;

export const TuiTemplates = {
  NoteList: async (ctx: { notes: Note[] }) => {
    return TuiRender(
      dedent(`
      {% if notes.length == 0 %}
      No notes found.
      {% else %}
      ## Notes

      {% for note in notes %}
      - [{{ note.file.relative }}](file://{{ note.file.filepath }})
      {% endfor %}

      {% endif %}
      `),
      { notes: ctx.notes }
    );
  },
};

export function createNoteService(options: {
  dbService: DbService;
  notebookPath?: string;
  configService: ConfigService;
}) {
  async function query(query: string) {
    const database = await options.dbService.getDb();

    const prepared = await database.prepare(query);
    const request = await prepared.send();
    const result = await request.readAll();

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
    Log.debug('readNote: filepath=%s', filepath);
    // const db = await options.dbService.getDb();
    //
    // try {
    //   const result = await db.query<{}>(`SELECT content, metadata FROM read_markdown('$filepath')`);
    //
    //   const rows = await result.getRowObjectsJson();
    //
    //   if (rows?.length === 0) {
    //     return null;
    //   }
    //   if (!rows[0]) {
    //     return null;
    //   }
    //
    //   return rows[0];
    // } catch {
    //   return null;
    // }
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
      SELECT * FROM read_markdown($filepath, include_filepath:=true)
    `);

    const request = await prepared.send({
      filepath: join(options.notebookPath, '**', '*.md'),
    });
    const rows = await request.readAll();

    if (!rows || rows.length === 0) {
      Log.debug('searchNotes: no results');
      return [];
    }

    const parsedRows = NoteDbSchema.array()(rows);
    if (parsedRows instanceof type.errors) {
      Log.error('searchNotes: failed to parse notes: %o', parsedRows);
      return [];
    }

    const parsed = NoteSchema.array()(
      parsedRows.map((note) => ({
        file: {
          filepath: note.file_path,
          relative: options.notebookPath
            ? note.file_path.replace(options.notebookPath + '/', '')
            : note.file_path,
        },
        content: note.content,
        metadata: note.metadata,
      }))
    );

    if (parsed instanceof type.errors) {
      Log.error('searchNotes: failed to parse notes: %s', prettifyArktypeErrors(parsed));
      return [];
    }

    Log.debug('searchNotes: found %d notes', parsed.length);

    return parsed;
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

  /**
   * Counts the number of markdown notes in the notebook path.
   * @returns The count of markdown notes
   */
  async function count(): Promise<number> {
    const db = await options.dbService.getDb();
    const notebookPath = options.notebookPath || '';
    const glob = join(notebookPath, '**', '*.md');
    Log.debug('count: notebookPath=%s glob=%s', notebookPath, glob);
    const prepared = await db.prepare(`SELECT COUNT(*) as count FROM read_markdown($pattern)`);
    Log.debug('count: prepared=%o', prepared);

    const rows = await prepared.send({ pattern: glob });

    const parsed = CountResultSchema(rows);
    if (parsed instanceof type.errors) {
      Log.error('count: failed to parse count result: %s', prettifyArktypeErrors(parsed));
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
