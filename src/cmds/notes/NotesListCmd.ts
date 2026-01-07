import { defineCommand } from 'clerc';
import { Logger } from '../../services/LoggerService';
import { requireNotebookMiddleware } from '../../middleware/requireNotebookMiddleware';
import { createNoteService } from '../../services/NoteService';

export const NotesListCommand = defineCommand(
  {
    name: 'notes list',
    description: 'List all notes in the project',
    flags: {},
    alias: [],
    parameters: [],
  },
  async (ctx) => {
    const notebook = await requireNotebookMiddleware({
      notebookService: ctx.store.notebooKService,
      path: ctx.flags.notebook,
    });

    if (!notebook) {
      return;
    }

    Logger.debug('NotesListCmd %s', notebook.config.path);

    const config = ctx.store.config?.store;
    const dbService = ctx.store.dbService;

    if (!config || !dbService) {
      // eslint-disable-next-line no-console
      console.error('Failed to load config or dbService');
      return;
    }

    const noteService = createNoteService({
      notebook,
      configService: config,
      dbService,
    });

    const results = await noteService.searchNotes();

    for (const note of results) {
      // eslint-disable-next-line no-console
      console.log(`- ${note.path}`);
    }
  }
);
