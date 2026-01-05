import type { Config, ConfigService } from './ConfigService';
import { promises as fs } from 'fs';
import { dirname, join } from 'path';
import { type } from 'arktype';
import { dedent, renderTemplateString, slugify } from '../core/strings';
import { Logger } from './LoggerService';
import { RenderMarkdownTui } from './Display';

const NotebookGroupSchema = type({
  name: 'string',
  description: 'string?',
  globs: 'string[]',
  metadata: type({ '[string]': 'string | number | boolean' }),
});

export type NotebookGroup = typeof NotebookGroupSchema.infer;

const NotebookConfigSchema = type({
  name: 'string',
  contexts: 'string[]?',
  templates: type({ '[string]': 'string' }).optional(),
  groups: NotebookGroupSchema.array().optional(),
});

type NotebookConfig = typeof NotebookConfigSchema.infer;

export interface Notebook {
  ready: Promise<void>;
  path: string;
  config: NotebookConfig;
  saveConfig(): Promise<void>;
  addContext(contextPath?: string): Promise<void>;
  loadTemplate(name: string): Promise<string | null>;
}

const TuiTemplates = {
  NotebookCreated: dedent(`
    # Notebook Created

    Your new notebook has been successfully created!

    - **Name**: {{name}}
    - **Path**: {{path}}

    You can start adding notes to your notebook right away.
  `),
  ContextAlreadyExists: dedent(`
    # Context Already Exists

    The context path is already associated with this notebook.

    - **Context**: {{contextPath}}
    - **Notebook**: {{notebookPath}}

    No changes were made.
  `),
  ContextAdded: dedent(`
    # Context Added

    The context path has been successfully added to your notebook.

    - **Context**: {{contextPath}}
    - **Notebook**: {{notebookPath}}

    This notebook will now be available when working in that directory.
  `),
  TemplateLoadError: dedent(`
    # Template Load Error

    Failed to load a template for your notebook. This may cause some features to be unavailable.

    - **Template Path**: {{templatePath}}
    - **Error**: {{error}}

    You may need to check the template file and try again.
  `),
};

export function createNotebookService(serviceOptions: { config: Config }) {
  class Notebook implements Notebook {
    static createNotebookConfigPath(path: string) {
      return join(path, `.${serviceOptions.config.configFilePath}`);
    }

    static async isNotebookPath(path?: string) {
      if (!path) {
        return false;
      }

      const configPath = Notebook.createNotebookConfigPath(path);
      if (!(await fs.exists(configPath))) {
        return false;
      }

      return true;
    }

    static async loadConfig(path: string): Promise<NotebookConfig | null> {
      const configPath = Notebook.createNotebookConfigPath(path);

      try {
        const content = await fs.readFile(configPath, 'utf-8');
        const parsed = JSON.parse(content);
        const config = NotebookConfigSchema(parsed);

        if (config instanceof type.errors) {
          Logger.debug('NotebookService.loadNotebookConfig: INVALID_CONFIG path=%s', configPath);
          return null;
        }

        return config;
      } catch (error) {
        const errorMsg = error instanceof Error ? error.message : String(error);
        Logger.debug(
          'NotebookService.loadNotebookConfig: ERROR path=%s error=%s',
          configPath,
          errorMsg
        );
        return null;
      }
    }

    /**
     * Initialize the notebook (load config, templates, etc)
     **/
    static async load(path: string): Promise<Notebook | null> {
      const config = await Notebook.loadConfig(path);
      if (!config) {
        return null;
      }
      return new Notebook(path, config);
    }

    static async new(args: { path: string; name: string }): Promise<Notebook> {
      const config: NotebookConfig = {
        name: args.name,
        templates: {},
        groups: [
          {
            name: 'Default',
            description: 'Default group for all notes',
            globs: ['**/*.md'],
            metadata: {},
          },
        ],
        contexts: [args.path],
      };

      const notebook = new Notebook(args.path, config);
      await notebook.saveConfig();

      RenderMarkdownTui(TuiTemplates.NotebookCreated, args);

      return notebook;
    }

    constructor(
      public path: string,
      public config: NotebookConfig
    ) {
      //
    }

    matchContext(path: string): string | null {
      if (!this.config.contexts) return null;
      return this.config.contexts.find((context) => path.startsWith(context)) || null;
    }

    /**
     * Write a notebook config to a given path
     */
    async saveConfig() {
      const configPath = Notebook.createNotebookConfigPath(this.path);
      try {
        const content = JSON.stringify(this.config, null, 2);
        await fs.mkdir(dirname(configPath), { recursive: true });
        await fs.writeFile(configPath, content, 'utf-8');
      } catch (error) {
        const errorMsg = error instanceof Error ? error.message : String(error);
        Logger.debug(
          'NotebookService.writeNotebookConfig: ERROR path=%s error=%s',
          configPath,
          errorMsg
        );
        throw error;
      }
    }

    /**
     * Add a path as a context for a notebook.
     *
     * Notebooks contexts are used to determine which notes belong to which notebooks.
     * A context is simply a path that is considered related to the notebook.
     */
    async addContext(contextPath: string = process.cwd()): Promise<void> {
      // Check if context already exists
      if (this.config.contexts?.includes(contextPath)) {
        await RenderMarkdownTui(TuiTemplates.ContextAlreadyExists, {
          contextPath,
          notebookPath: this.path,
        });

        return;
      }

      // Add the context
      this.config.contexts = [...(this.config.contexts || []), contextPath];
      await this.saveConfig();

      await RenderMarkdownTui(TuiTemplates.ContextAdded, {
        contextPath,
        notebookPath: this.path,
      });
    }

    async loadTemplate(name: string): Promise<string | null> {
      if (!this.config.templates) {
        return null;
      }
      const templatePath = this.config.templates[name];
      if (!templatePath) {
        return null;
      }

      // Load templates
      // templates are listed as a mapping of template name to file path
      try {
        const template = await import(templatePath, { assert: { type: 'markdown' } }).then(
          (mod) => mod.default
        );
        return template;
      } catch (error) {
        const errorMsg = error instanceof Error ? error.message : String(error);
        Logger.debug(
          'NotebookService.getNotebook: ERROR_LOADING_TEMPLATE path=%s error=%s',
          templatePath,
          errorMsg
        );
        await RenderMarkdownTui(TuiTemplates.TemplateLoadError, {
          templatePath,
          error: errorMsg,
        });
      }
      return null;
    }
  }

  /**
   * Get a notebook by its path
   */
  async function open(notebookPath: string): Promise<Notebook | null> {
    return Notebook.load(notebookPath);
  }

  async function create(args: { name: string; path?: string }): Promise<Notebook> {
    const notebookPath = args.path || process.cwd();
    return Notebook.new({ name: args.name, path: notebookPath });
  }

  /**
   * Discover the notebook path based on the current working directory.
   *
   * A notebook path is any folder that contains a .wiki/config.json
   *
   * Priority:
   *
   *  1. Declared Notebook Path
   *  2. Context Matching in Notebook Configs
   *  3. Ancestor Directory Search
   *
   * @param cwd Current working directory (defaults to process.cwd())
   * @returns Resolved notebook path or null if not found
   */
  async function infer(cwd: string = process.cwd()): Promise<Notebook | null> {
    const notebookPath = serviceOptions.config.notebookPath;
    // Step 1: Check environment/cli-arg variable (resolved and provided by the ConfigService)
    if (notebookPath && (await Notebook.isNotebookPath(notebookPath))) {
      const notebook = await Notebook.load(notebookPath);
      if (notebook) {
        Logger.debug('NotebookService.discoverNotebookPath: USE_DECLARED_PATH %s', notebookPath);
        return notebook;
      }
    }

    for (const notebook of await list(cwd)) {
      if (!notebook.matchContext(cwd)) {
        continue;
      }

      Logger.debug(
        'NotebookService.discoverNotebookPath: MATCHED_LISTED_NOTEBOOK %s',
        notebook.path
      );
      return notebook;
    }

    Logger.debug('NotebookService.discoverNotebookPath: NO_NOTEBOOK_FOUND');
    return null;
  }

  async function list(cwd: string = process.cwd()): Promise<Notebook[]> {
    const output: Notebook[] = [];

    // STEP 2: Check for notebook configs in config.notebooks
    for (const notebookPath of serviceOptions.config.notebooks) {
      const configFilePath = await Notebook.isNotebookPath(notebookPath);
      if (!configFilePath) {
        continue;
      }

      const notebook = await Notebook.load(notebookPath);
      if (!notebook) {
        continue;
      }
      output.push(notebook);
    }

    // Step 3: Search ancestor directories
    let next = cwd;
    while (next !== '') {
      const configFilePath = await Notebook.isNotebookPath(next);
      const notebookPath = next;
      next = dirname(next);

      if (!configFilePath) {
        continue;
      }

      const notebook = await Notebook.load(notebookPath);
      if (!notebook) {
        continue;
      }

      output.push(notebook);
    }

    return output;
  }

  /**
   * Return the public API
   */
  return {
    list,
    create,
    open,
    infer,
  };
}

export type NotebookService = ReturnType<typeof createNotebookService>;
