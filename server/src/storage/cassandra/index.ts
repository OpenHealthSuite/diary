import { Client, DseClientOptions } from 'cassandra-driver'

export * from './FoodLogStorageFunctions'
export * from './migrations'

export const DEFAULT_CLIENT_CONFIG = {
    contactPoints: process.env.OPENFOODDIARY_CASSANDRA_CONTACT_POINTS ? 
      process.env.OPENFOODDIARY_CASSANDRA_CONTACT_POINTS.split(';') : ['localhost:9042'],
    localDataCenter: process.env.OPENFOODDIARY_CASSANDRA_LOCALDATACENTER ?? 'datacenter1',
    credentials: {
      username: process.env.OPENFOODDIARY_CASSANDRA_USER ?? 'cassandra',
      password: process.env.OPENFOODDIARY_CASSANDRA_PASSWORD ?? 'cassandra'
    }
}

export function createClient(clientConfig: DseClientOptions = DEFAULT_CLIENT_CONFIG) {
    return new Client(clientConfig);
}

export function setGlobalClient(client: Client) {
  CASSANDRA_CLIENT = client;
}

export function setDefaultGlobalClient() {
  CASSANDRA_CLIENT = createClient(DEFAULT_CLIENT_CONFIG);
}

export let CASSANDRA_CLIENT: Client;

export async function closeCassandra() {
    await CASSANDRA_CLIENT.shutdown();
}