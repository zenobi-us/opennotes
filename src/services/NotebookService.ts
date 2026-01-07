import { NotebookConfigFile, type ConfigService } from './ConfigService';
import { promises as fs } from 'fs';
import { dirname, join, resolve } from 'path';
import { type } from 'arktype';
import { dedent } from '../core/strings';
import { Logger } from './LoggerService';
import { TuiRender } from './Display';
import { createNoteService, type NoteService } from './NoteService';
import type { DbService } from './Db';

const Log = Logger.child({ namespace: 'NotebookService' });

const NotebookGroupSchema = type({
  /**
   * The name of the notebook group
   **/
  name: 'string',
  /**
   * Notes created in these globs will belong to this group
   **/
  globs: 'string[]',
  /**
   * Metadata that notes created here will automatically have
   **/
  metadata: type({ '[string]': 'string | number | boolean' }),
  /**
   * Template used for notes created in this group
   **/
  template: 'string?',
});

export type NotebookGroup = typeof NotebookGroupSchema.infer;

const NotebookConfigSchema = type({
  /**
   * The root of where notes are stored for this notebook
   **/
  path: 'string',
  /**
   * The display name of the notebook
   **/
  name: 'string',
  /**
   * The context paths associated with this notebook
   *
   * Context paths are used to determine which notes belong to which notebooks.
   * A context is simply a path that is considered related to the notebook.
   **/
  contexts: 'string[]?',
  /**
   * Templates associated with this notebook
   *
   * This is a mapping of template name to file path
   **/
  templates: type({ '[string]': 'string' }).optional(),
  /**
   * Groups of notes within the notebook
   **/
  groups: NotebookGroupSchema.array().optional(),
});

type NotebookConfig = typeof NotebookConfigSchema.infer;

export interface INotebook {
  /**
   * The notebook configuration
   *
   * This is loaded fron the notebook's config.json file
   */
  config: NotebookConfig;

  /**
   * A NotesService bound to this notebook
   */
  notes: NoteService;

  /**
   * Save the notebook configuration to disk
   */
  saveConfig(): Promise<void>;

  /**
   * Add a context path to the notebook
   */
  addContext(_contextPath?: string): Promise<void>;

  /**
   * Match a context path to see if it belongs to this notebook
   */
  matchContext(path: string): string | null;

  /**
   * Load a template by name
   */
  loadTemplate(_name: string): Promise<string | null>;
}

export const TuiTemplates = {
  NotebookCreated: (ctx: { notebook: INotebook }) =>
    TuiRender(
      dedent(`
    # Notebook Created

    Your new notebook has been successfully created!

    - **Name**: {{notebook.config.name}}
    - **Path**: {{notebook.config.path}}

    You can start adding notes to your notebook right away.
  `),
      ctx
    ),
  ContextAlreadyExists: (ctx: { path: string; notebook: INotebook }) =>
    TuiRender(
      dedent(`
    # Context Already Exists

    The context path is already associated with this notebook.

    - **Context**: {{path}}
    - **Notebook**: {{notebook.config.path}}

    No changes were made.
  `),
      ctx
    ),
  ContextAdded: (ctx: { path: string; notebook: INotebook }) =>
    TuiRender(
      dedent(`
    # Context Added

    The context path has been successfully added to your notebook.

    - **Context**: {{path}}
    - **Notebook**: {{notebook.config.path}}

    This notebook will now be available when working in that directory.
  `),
      ctx
    ),
  TemplateLoadError: (ctx: { templatePath: string; error: string }) =>
    TuiRender(
      dedent(`
    # Template Load Error

    Failed to load a template for your notebook. This may cause some features to be unavailable.

    - **Template Path**: {{templatePath}}
    - **Error**: {{error}}

    You may need to check the template file and try again.
  `),
      ctx
    ),
  NotebookInfo: async (ctx: { notebook: INotebook }) => {
    const count = await ctx.notebook.notes.count();
    return TuiRender(
      dedent(`
    # Notebook Information

    - **Name**: {{notebook.config.name}}
    - **Path**: {{notebook.config.path}}
    - **Notes**: {{ noteCount }}

    ## Contexts

    {% for context in notebook.config.contexts %}
      - {{ context }}
    {% empty %}
      - _No contexts defined_
    {% endfor %}

    ## Groups

    {% for group in notebook.config.groups %}
      - **Name**: {{ group.name }}
        **Template**: {{ group.template | default('_No template defined_') }}
        **Metadata**:
        {% for key, value in group.metadata %}
          - {{ key }}: {{ value }}
        {% empty %}
          - _No metadata defined_
        {% endfor %}
        - Globs:
        {% for glob in group.globs %}
          - {{ glob }}
        {% empty %}
          - _No globs defined_
        {% endfor %}
    {% endfor %}
  `),
      {
        notebook: ctx.notebook,
        noteCount: count,
      }
    );
  },
  CreateYourFirstNotebook: () =>
    TuiRender(
      dedent(`
    # No Notebooks Found

    It looks like you don't have any notebooks set up yet.

    To create your first notebook, use the following command:

    \`\`\`bash
    wiki notebook create --name "My First Notebook"
    \`\`\`

    This will create a new notebook in your current directory.
  `),
      {}
    ),
  DisplayNotebookList: (ctx: { notebooks?: INotebook[] }) =>
    TuiRender(
      dedent(`
    {% if notebooks.length > 0 %}
    {% for notebook in notebooks %}
    - **{{notebook.config.name}}**: {{notebook.path}}
    {% endfor %}

    {% else %}
    No notebooks are currently registered.

    To create a new notebook, use the command:

    \`\`\`bash
    wiki notebook create --name "My Notebook Name" --global
    \`\`\`

    {% endif %}
  `),
      ctx
    ),
};

export function createNotebookService(serviceOptions: {
  configService: ConfigService;
  dbService: DbService;
}) {
  class Notebook implements INotebook {
    static createNotebookConfigPath(path: string) {
      return join(path, NotebookConfigFile);
    }

    /**
     * Check if a given path is a notebook path
     *
     * @see {NotebookConfigFile}
     */
    static async hasNotebook(path?: string) {
      if (!path) {
        return false;
      }

      const configPath = Notebook.createNotebookConfigPath(path);
      if (await fs.exists(configPath)) {
        return true;
      }

      return false;
    }

    static async loadConfig(path: string): Promise<NotebookConfig | null> {
      Log.debug('Notebook.loadConfig: path=%s', path);
      const configPath = Notebook.createNotebookConfigPath(path);

      try {
        const content = await fs.readFile(configPath, 'utf-8');
        const parsed = JSON.parse(content);
        const config = NotebookConfigSchema(parsed);

        if (config instanceof type.errors) {
          Log.error('NotebookService.loadNotebookConfig: INVALID_CONFIG path=%s', configPath);
          return null;
        }

        return config;
      } catch (error) {
        const errorMsg = error instanceof Error ? error.message : String(error);
        Log.error(
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
      const noteService = createNoteService({
        configService: serviceOptions.configService,
        notebookPath: resolve(path, config.path),
        dbService: serviceOptions.dbService,
      });

      return new Notebook(config, noteService);
    }

    static async new(args: { path: string; name: string }): Promise<Notebook> {
      const config: NotebookConfig = {
        name: args.name,
        /*
         * Store notes relative to the notebook path in a .notes folder
         * This keeps notes organized and separate from other project files.
         */
        path: join('.', '.notes'),
        templates: {},
        groups: [
          {
            name: 'Default',
            globs: [join('**', '*.md')],
            metadata: {},
          },
        ],
        contexts: [args.path],
      };
      Log.debug('Notebook.new: path=%s name=%s', config.path, config.name);

      const noteService = createNoteService({
        configService: serviceOptions.configService,
        notebookPath: join(args.path, '.notes'),
        dbService: serviceOptions.dbService,
      });

      const notebook = new Notebook(config, noteService);
      await notebook.saveConfig();

      await TuiTemplates.NotebookCreated({
        notebook,
      });

      return notebook;
    }

    constructor(
      public config: NotebookConfig,
      public notes: NoteService
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
      const config = NotebookConfigSchema(this.config);
      if (config instanceof type.errors) {
        throw new Error(`Invalid notebook config: ${config.toString()}`);
      }

      const configPath = dirname(config.path);
      try {
        const content = JSON.stringify(this.config, null, 2);
        await fs.mkdir(dirname(configPath), { recursive: true });
        await fs.writeFile(configPath, content, 'utf-8');
      } catch (error) {
        const errorMsg = error instanceof Error ? error.message : String(error);
        Log.debug(
          'NotebookService.writeNotebookConfig: ERROR path=%s error=%s',
          configPath,
          errorMsg
        );
        throw error;
      }
    }

    /**
     * Add a context path to the notebook.
     *
     * Notebooks contexts are used to determine which notes belong to which notebooks.
     * A context is simply a path that is considered related to the notebook.
     */
    async addContext(_contextPath: string = process.cwd()): Promise<void> {
      // Check if context already exists
      if (this.config.contexts?.includes(_contextPath)) {
        await TuiTemplates.ContextAlreadyExists({
          path: _contextPath,
          notebook: this,
        });

        return;
      }

      // Add the context
      this.config.contexts = [...(this.config.contexts || []), _contextPath];
      await this.saveConfig();

      await TuiTemplates.ContextAdded({
        path: _contextPath,
        notebook: this,
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
        Log.debug(
          'NotebookService.getNotebook: ERROR_LOADING_TEMPLATE path=%s error=%s',
          templatePath,
          errorMsg
        );
        await TuiTemplates.TemplateLoadError({
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
    return Notebook.new({
      name: args.name,
      path: args.path || process.cwd(),
    });
  }

  /**
   * Discover the notebook path based on the current working directory.
   *
   * A notebook path is any folder that contains a .opennotes/config.json
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
    Log.debug('NotebookService.infer: cwd=%s', cwd);
    const notebookPath = serviceOptions.configService.store.notebookPath;
    // Step 1: Check environment/cli-arg variable (resolved and provided by the ConfigService)
    if (notebookPath && (await Notebook.hasNotebook(notebookPath))) {
      const notebook = await Notebook.load(notebookPath);
      if (notebook) {
        Log.debug('Notebook.infer: USE_DECLARED_PATH %s', notebookPath);
        return notebook;
      }
    }

    for (const notebook of await list(cwd)) {
      if (!notebook.matchContext(cwd)) {
        continue;
      }

      Log.debug('Notebook.infer: MATCHED_LISTED_NOTEBOOK %s', notebook.config.path);
      return notebook;
    }

    Log.debug('Notebook.infer: NO_NOTEBOOK_FOUND');
    return null;
  }

  async function list(cwd: string = process.cwd()): Promise<Notebook[]> {
    Log.debug('list: cwd=%s', cwd);
    const registered_notebooks: Notebook[] = [];

    // STEP 2: Check for notebook configs in config.notebooks
    for (const notebookPath of serviceOptions.configService.store.notebooks) {
      const configFilePath = await Notebook.hasNotebook(notebookPath);
      if (!configFilePath) {
        continue;
      }
      Log.debug('list.AttemptLoadNotebook: %s', notebookPath);
      const notebook = await Notebook.load(notebookPath);
      if (!notebook) {
        continue;
      }
      registered_notebooks.push(notebook);
    }

    Log.debug('list: found %d notebooks from config', registered_notebooks.length);

    const ancestor_notebooks: Notebook[] = [];
    // Step 3: Search ancestor directories
    let next = cwd;
    while (next !== '/') {
      const configFilePath = await Notebook.hasNotebook(next);
      const notebookPath = next;
      next = dirname(next);

      if (!configFilePath) {
        continue;
      }
      Log.debug('list.AttemptLoadAncestorNotebook: %s', notebookPath);

      const notebook = await Notebook.load(notebookPath);
      if (!notebook) {
        continue;
      }

      ancestor_notebooks.push(notebook);
    }
    Log.debug('list: found %d ancestor notebooks', ancestor_notebooks.length);

    return [...registered_notebooks, ...ancestor_notebooks];
  }

  /**
   * Get information about a notebook
   */
  async function info(args: { notebook: Notebook } | { notebookPath: string }) {
    let notebook: Notebook | null = null;
    if ('notebook' in args) {
      notebook = args.notebook;
    } else if ('notebookPath' in args) {
      notebook = await open(args.notebookPath);
    }

    if (!notebook) {
      Log.error('info: NOTEBOOK_NOT_FOUND');
      return null;
    }

    const content = await TuiTemplates.NotebookInfo({ notebook });
    // eslint-disable-next-line no-console
    console.log(content);
  }

  Log.debug('Ready');
  /**
   * Return the public API
   */
  return {
    list,
    create,
    open,
    info,
    infer,
  };
}

export type NotebookService = ReturnType<typeof createNotebookService>;
