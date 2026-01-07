import type { Config as ConfigShape } from 'bunfig';
import { loadConfig } from 'bunfig';
import { type } from 'arktype';
import { promises as fs } from 'fs';
import { join, dirname } from 'node:path';
import envPaths from 'env-paths';
import { Logger } from './LoggerService';

const Log = Logger.child({ namespace: 'ConfigService' });

const Paths = envPaths('opennotes', { suffix: '' });

/**
 * Config File found in the user's config directory
 * We consider this to be the global config
 */
export const UserConfigFile = join(Paths.config, 'config.json');
/**
 * Config File
 *
 * These look like:
 *
 * .opennotes.json (path: ./SomeNotebook)
 * SomeNotebook/
 *   SomeNote.md
 *   AnotherNote.md
 *   AFolder/
 *     NestedNote.md
 *     etc.md
 *
 * or
 *
 * AnotherNotebook/
 *   .opennotes.json (path: ./)
 *   Notes.md
 */
export const NotebookConfigFile = '.opennotes.json';

export const ConfigSchema = type({
  /**
   * Notebook paths are directories where there is a NotebookConfigFile
   */
  notebooks: 'string[]',
  /**
   * Current notebook path
   *
   * This can be provided from either (first one found wins):
   * - OPENNOTES_NOTEBOOK_PATH env variable
   * - --notebookPath CLI flag
   * - or stored in global config.
   */
  notebookPath: 'string?',
});

export type Config = typeof ConfigSchema.infer;

type ConfigWriter = (config: Config) => Promise<void>;

export type ConfigService = {
  store: Config;
  write: ConfigWriter;
};

const options: ConfigShape<Config> = {
  name: 'opentask',
  cwd: './',
  defaultConfig: {
    notebooks: [join(Paths.config, 'notebooks')],
  },
};

export async function createConfigService(): Promise<ConfigService> {
  Log.debug('Loading');
  const store = await loadConfig(options);
  Log.debug('Loadeded %o', { store });

  async function write(config: Config): Promise<void> {
    Log.debug('Writing: %o', { config });
    await fs.mkdir(dirname(UserConfigFile), { recursive: true });
    await fs.writeFile(UserConfigFile, JSON.stringify(config, null, 2));
    Log.debug('Written');
  }

  Log.debug('Ready');

  return {
    store,
    write,
  };
}
