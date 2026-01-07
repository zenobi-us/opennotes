import * as duckdb from '@duckdb/duckdb-wasm';
import { Logger } from './LoggerService';

const Log = Logger.child({ namespace: 'DbService' });

export type DbService = ReturnType<typeof createDbService>;

const bundles = {
  mainModule: import.meta.resolve('@duckdb/duckdb-wasm/dist/duckdb-mvp.wasm'),
  mainWorker: import.meta.resolve('@duckdb/duckdb-wasm/dist/duckdb-browser-mvp.worker.js'),
};

export function createDbService() {
  let db: duckdb.AsyncDuckDB | null = null;

  async function getDb(): Promise<duckdb.AsyncDuckDBConnection> {
    if (db !== null) {
      return await db.connect();
    }

    // eslint-disable-next-line no-undef
    const worker = new Worker(bundles.mainWorker);
    const logger = new duckdb.ConsoleLogger();

    const instance = new duckdb.AsyncDuckDB(logger, worker);
    await instance.instantiate(bundles.mainModule);

    const connection = await instance.connect();

    try {
      // Install and load the markdown extension
      connection.query('INSTALL markdown FROM community;');
      connection.query('LOAD markdown;');
    } catch (error) {
      Log.error('Failed to initialize markdown extension: %s', error);
      throw new Error(
        `Failed to initialize markdown extension: ${error instanceof Error ? error.message : String(error)}`
      );
    }

    return connection;
  }

  Log.debug('initialized');
  return {
    getDb,
  };
}
