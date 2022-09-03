export * from './types/StorageErrors';
import { FoodLogStorage } from './interfaces';
import * as cassandraFoodLogStorage from './cassandra/FoodLogStorageFunctions'

export const cassandra: { foodLog: FoodLogStorage } = {
    foodLog: cassandraFoodLogStorage
}