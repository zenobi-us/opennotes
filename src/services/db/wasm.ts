import { Logger } from '../LoggerService';
import { DuckDB } from '@hpcc-js/wasm-duckdb';

const Log = Logger.child({ namespace: 'DbService/Wasm' });

export type DbService = ReturnType<typeof createDbService>;

type DbConnection = Awaited<ReturnType<DuckDB['db']['connect']>>;

export function createDbService() {
  let connection: DbConnection | null = null;

  async function getDb() {
    if (connection) {
      Log.debug('getDb: return existing instance');
      return connection;
    }

    Log.debug('getDb: initialize new instance');
    const duckdb = await DuckDB.load();
    connection = await duckdb.db.connect();

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
