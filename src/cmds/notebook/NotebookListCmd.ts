import { defineCommand } from 'clerc';
import { Logger } from '../../services/LoggerService';
import { TuiTemplates as NotebookTuiTemplates } from '../../services/NotebookService';

const Log = Logger.child({ namespace: 'NotebookListCmd' });

export const NotebookListCommand = defineCommand(
  {
    name: 'notebook list',
    description: 'Manage wiki notebooks',
    flags: {},
    alias: ['nb list'],
    parameters: [],
  },
  async (ctx) => {
    Log.debug('Execute');
    const notebooks = await ctx.store.notebooKService?.list();

    // eslint-disable-next-line no-console
    console.log(await NotebookTuiTemplates.DisplayNotebookList({ notebooks }));
  }
);
