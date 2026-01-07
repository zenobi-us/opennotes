import {
  TuiTemplates as NotebookTuiTemplates,
  type NotebookService,
} from '../services/NotebookService';

export async function requireNotebookMiddleware(args: {
  path?: string;
  notebookService?: NotebookService;
}) {
  let notebook = args?.path
    ? await args.notebookService?.open(args.path)
    : await args.notebookService?.infer();

  if (!notebook) {
    // eslint-disable-next-line no-console
    console.error(await NotebookTuiTemplates.CreateYourFirstNotebook());
    return null;
  }

  return notebook;
}
