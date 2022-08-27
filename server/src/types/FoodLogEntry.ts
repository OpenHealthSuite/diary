export type FoodLogEntry = {
    id: string,
    name: string,
    labels: Set<string>,
    time: {
        start: Date,
        end: Date
    },
    metrics: {
        [key: string]: number
    }
}

export type CreateFoodLogEntry = Pick<FoodLogEntry, "name" | "labels" | "time" | "metrics" >

export type EditFoodLogEntry = Pick<FoodLogEntry, "id" | "name" | "labels" | "time" | "metrics" >