import { Client, DseClientOptions } from 'cassandra-driver'
import { createClient, DEFAULT_CLIENT_CONFIG } from '.'

// Probably worth looking into this more formally
const MIGRATIONS: string[] = [
    "CREATE KEYSPACE IF NOT EXISTS openfooddiary WITH REPLICATION = {'class':'SimpleStrategy','replication_factor':1};", // TODO: Make this optional?
    `CREATE TYPE IF NOT EXISTS openfooddiary.logTimes (start timestamp, end timestamp);`,
    `CREATE TABLE IF NOT EXISTS openfooddiary.user_foodlogentry (userId UUID, id UUID, name text, labels set<text>,time frozen<openfooddiary.logTimes>,metrics map<text,int>, timeStart timestamp, timeEnd timestamp, PRIMARY KEY ((userId), id));`,
    `CREATE INDEX IF NOT EXISTS ON openfooddiary.user_foodlogentry (timeStart);`,
    `CREATE INDEX IF NOT EXISTS ON openfooddiary.user_foodlogentry (timeEnd);`,
]

export async function doCassandraMigrations(clientConfig: DseClientOptions = DEFAULT_CLIENT_CONFIG) {
    let migrationClient = createClient(clientConfig)
    await migrationClient.connect().then(async () => {
        for (let migration of MIGRATIONS) {
            await migrationClient.execute(migration)
        }
        return migrationClient.shutdown();
    })
}