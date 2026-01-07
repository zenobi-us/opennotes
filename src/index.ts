import { Cli } from 'clerc';
import { friendlyErrorPlugin } from '@clerc/plugin-friendly-error';
import { notFoundPlugin } from '@clerc/plugin-not-found';
import { strictFlagsPlugin } from '@clerc/plugin-strict-flags';
import { updateNotifierPlugin } from '@clerc/plugin-update-notifier';

import pkg from '../package.json' assert { type: 'json' };

import { getGitTag } from './macros/GitInfo.ts' with { type: 'macro' };
import { InitCommand } from './cmds/init/InitCmd.ts';
import { NotebookCommand } from './cmds/notebook/NotebookCmd.ts';
import { NotebookListCommand } from './cmds/notebook/NotebookListCmd.ts';
import { NotebookAddContextPathCommand } from './cmds/notebook/NotebookAddContextPathCmd.ts';
import { NotesCommand } from './cmds/notes/NotesCmd.ts';
import { NotesAddCommand } from './cmds/notes/NotesAddCmd.ts';
import { NotesListCommand } from './cmds/notes/NotesListCmd.ts';
import { NotesRemoveCommand } from './cmds/notes/NotesRemoveCmd.ts';
import { NotesSearchCommand } from './cmds/notes/NotesSearchCmd.ts';
import { Logger } from './services/LoggerService.ts';
import { createConfigService } from './services/ConfigService.ts';
import { createNotebookService } from './services/NotebookService.ts';

import type { ConfigService } from './services/ConfigService.ts';
import type { NotebookService } from './services/NotebookService.ts';
import { NotebookCreateCommand } from './cmds/notebook/NotebookCreateCmd.ts';
import { createDbService, type DbService } from './services/Db.ts';

declare module '@clerc/core' {
  export interface ContextStore {
    config: ConfigService;
    notebooKService: NotebookService;
    dbService: DbService;
  }
}

const Log = Logger.child({ namespace: 'CLI' });

Cli() // Create a new CLI with help and version plugins
  .name('wiki') // Optional, CLI readable name
  .scriptName('wiki') // CLI script name (the command used to run the CLI)
  .description('A wiki CLI') // CLI description
  .version(getGitTag() || 'dev') // CLI version
  .use(friendlyErrorPlugin()) // use the friendly error plugin to handle errors gracefully
  .use(notFoundPlugin()) // use the not found plugin to handle unknown commands
  .use(strictFlagsPlugin()) // use the strict flags plugin to enforce strict flag parsing
  .use(
    updateNotifierPlugin({
      notify: {},
      // @ts-expect-error pkg is json
      pkg,
    })
  ) // use the update notifier plugin to notify users of updates
  .interceptor(async (ctx, next) => {
    Log.debug('Interceptor.before');

    const dbService = createDbService();
    const configService = await createConfigService();

    const notebookService = createNotebookService({ configService, dbService });

    ctx.store.config = configService;
    ctx.store.notebooKService = notebookService;
    ctx.store.dbService = dbService;

    await next();
    Log.debug('Interceptor.after');
  })
  .command([
    InitCommand,
    NotebookCommand,
    NotebookListCommand,
    NotebookAddContextPathCommand,
    NotebookCreateCommand,
    NotesCommand,
    NotesAddCommand,
    NotesListCommand,
    NotesRemoveCommand,
    NotesSearchCommand,
  ]) // register the init, notebook and notes commands
  .parse(); // Parse the CLI arguments and execute commands
