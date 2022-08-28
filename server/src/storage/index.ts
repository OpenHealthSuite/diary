export * from './types/StorageErrors';
import { FoodLogStorage } from './interfaces';
import * as memoryFoodLogStorage from './memory/FoodLogStorageFunctions'

export const memory: { foodLog: FoodLogStorage } = {
    foodLog: memoryFoodLogStorage
}