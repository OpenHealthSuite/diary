import { DeleteFoodLogFunction, EditFoodLogFunction, RetrieveFoodLogFunction, StoreFoodLogFunction } from "./types";

export interface FoodLogStorage {
    storeFoodLog: StoreFoodLogFunction,
    retrieveFoodLog: RetrieveFoodLogFunction,
    editFoodLog: EditFoodLogFunction,
    deleteFoodLog: DeleteFoodLogFunction
}