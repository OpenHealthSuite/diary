import { DeleteFoodLogFunction, EditFoodLogFunction, QueryFoodLogFunction, RetrieveFoodLogFunction, StoreFoodLogFunction } from "./types";

export interface FoodLogStorage {
    storeFoodLog: StoreFoodLogFunction,
    retrieveFoodLog: RetrieveFoodLogFunction,
    editFoodLog: EditFoodLogFunction,
    deleteFoodLog: DeleteFoodLogFunction,
    queryFoodLogs: QueryFoodLogFunction
}