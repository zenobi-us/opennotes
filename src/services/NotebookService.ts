import { NotebookConfigFile, type ConfigService } from './ConfigService';
import { promises as fs } from 'fs';
import { dirname, join, relative, resolve } from 'path';
import { type } from 'arktype';
import { dedent } from '../core/strings';
import { Logger } from './LoggerService';
import { TuiRender } from './Display';
import { createNoteService, type NoteService } from './NoteService';
import type { DbService } from './Db';
import { prettifyArktypeErrors } from '../core/schema';

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

const StoredNotebookConfigSchema = type({
  /**
   * The root of where notes are stored for this notebook
   **/
  root: 'string',
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

const NotebookConfigSchema = type({
  /**
   * The path to the notebook config file
   **/
  path: 'string',
}).merge(StoredNotebookConfigSchema);

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
  saveConfig(args?: { register?: boolean }): Promise<void>;

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
  NotebookRegistered: (ctx: { notebook: INotebook; path: string }) =>
    TuiRender(
      dedent(`
     # Notebook Registered

     Your notebook has been successfully registered globally!

     - **Name**: {{notebook.config.name}}
     - **Path**: {{path}}

     This notebook is now available from anywhere on your system.
   `),
      ctx
    ),
  InvalidNotebookConfig: (ctx: { path: string; error: string }) =>
    TuiRender(
      dedent(`
     # Invalid Notebook Configuration

     The notebook at the specified path has an invalid configuration.

     - **Path**: {{path}}
     - **Error**: {{error}}

     Please fix the .opennotes.json file and try again. You can delete it and create a new notebook if needed.
   `),
      ctx
    ),
  NotebookCreated: (ctx: { notebook: INotebook }) =>
    TuiRender(
      dedent(`
     # Notebook Created

     Your new notebook has been successfully created!

     - **Name**: {{notebook.config.name}}
     - **Path**: {{notebook.config.root}}
     - **Config Path**: {{notebook.config.path}}

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
    - **Notebook**: {{notebook.config.root}}
    - **Config Path**: {{notebook.config.path}}

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
    - **Notebook**: {{notebook.config.root}}
    - **Config Path**: {{notebook.config.path}}

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
  NotebookInfo: async (ctx: {
    notebook: INotebook;
    notes: {
      count: number;
    };
  }) => {
    return TuiRender(
      dedent(`
    # Notebook Information

    - **Name**: {{notebook.config.name}}
    - **Path**: {{notebook.config.root}}
    - **Config Path**: {{notebook.config.path}}
    - **Notes**: {{ notes.count }}

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
      ctx
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
        Log.debug('Notebook.loadConfig: read config %s. %s', configPath, content);
        const parsed = JSON.parse(content);
        Log.debug('Notebook.loadConfig: parsed config %o', parsed);
        const storedConfig = StoredNotebookConfigSchema(parsed);

        if (storedConfig instanceof type.errors) {
          Log.error(
            'NotebookService.loadNotebookConfig: INVALID_CONFIG path=%s. \n %s',
            configPath,
            prettifyArktypeErrors(storedConfig)
          );
          return null;
        }

        // verify if the declared config.path exists
        const notebookNotesPath = resolve(path, storedConfig.root);
        if (!(await fs.exists(notebookNotesPath))) {
          Log.error(
            'NotebookService.loadNotebookConfig: NOTES_PATH_NOT_FOUND path=%s',
            notebookNotesPath
          );
          return null;
        }

        const config = NotebookConfigSchema({
          ...storedConfig,
          root: notebookNotesPath,
          path: configPath,
        });

        if (config instanceof type.errors) {
          Log.error(
            'NotebookService.loadNotebookConfig: INVALID_CONFIG_AFTER_RESOLVE path=%s. \n %s',
            configPath,
            prettifyArktypeErrors(config)
          );
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
        notebookPath: resolve(path, config.root),
        dbService: serviceOptions.dbService,
      });

      return new Notebook(config, noteService);
    }

    static async new(args: { path: string; name: string; register?: boolean }): Promise<Notebook> {
      const config = NotebookConfigSchema({
        name: args.name,
        /*
         * Store notes relative to the notebook path in a .notes folder
         * This keeps notes organized and separate from other project files.
         */
        root: join('.', '.notes'),
        path: Notebook.createNotebookConfigPath(args.path),
        templates: {},
        groups: [
          {
            name: 'Default',
            globs: [join('**', '*.md')],
            metadata: {},
          },
        ],
        contexts: [args.path],
      });

      if (config instanceof type.errors) {
        Log.error(
          'Notebook.new: INVALID_CONFIG path=%s. \n %s',
          args.path,
          prettifyArktypeErrors(config)
        );
        throw new Error(`Invalid notebook config: ${prettifyArktypeErrors(config)}`);
      }

      Log.debug('Notebook.new: path=%s name=%s', config.root, config.name);

      const noteService = createNoteService({
        configService: serviceOptions.configService,
        notebookPath: join(args.path, '.notes'),
        dbService: serviceOptions.dbService,
      });

      const notebook = new Notebook(config, noteService);
      await notebook.saveConfig({ register: args.register });

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
    async saveConfig(args?: { register?: boolean }): Promise<void> {
      const config = NotebookConfigSchema(this.config);
      if (config instanceof type.errors) {
        throw new Error(`Invalid notebook config: ${config.toString()}`);
      }

      const { path, root, ...configToSave } = config;
      const configPath = path;
      const rootDir = relative(dirname(configPath), root);

      try {
        const content = JSON.stringify(
          { ...configToSave, root: rootDir === '' ? '.' : rootDir },
          null,
          2
        );
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

      // Register the notebook globally if requested
      if (args?.register) {
        const notebookPath = this.config.root;
        const notebooks = serviceOptions.configService.store.notebooks || [];
        if (notebooks.includes(notebookPath)) {
          return;
        }

        notebooks.push(notebookPath);
        await serviceOptions.configService.write({
          ...serviceOptions.configService.store,
          notebooks,
        });
        Log.debug('Notebook registered globally: %s', notebookPath);
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

  async function create(args: {
    name: string;
    path?: string;
    register?: boolean;
  }): Promise<Notebook> {
    const notebookPath = args.path || process.cwd();
    const notebook = await Notebook.new({
      name: args.name,
      path: notebookPath,
      register: args.register,
    });

    return notebook;
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

      Log.debug('Notebook.infer: MATCHED_LISTED_NOTEBOOK %s', notebook.config.root);
      return notebook;
    }

    Log.debug('Notebook.infer: NO_NOTEBOOK_FOUND');
    return null;
  }

  async function list(cwd: string = process.cwd()): Promise<Notebook[]> {
    Log.debug('list: cwd=%s', cwd);
    const registered_notebooks: Notebook[] = [];
    Log.debug('list.notebooks. config: %o', serviceOptions.configService.store);

    // STEP 2: Check for notebook configs in config.notebooks
    for (const notebookPath of serviceOptions.configService.store.notebooks) {
      Log.debug('list.CheckRegisteredNotebook: %s', notebookPath);
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

    const notes = {
      count: await notebook.notes.count(),
    };

    const content = await TuiTemplates.NotebookInfo({ notebook, notes });
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
