export type FoodLogEntry = {
    userId: string,
    id: string,
    name: string,
    labels: Set<string>,
    time: {
        start: Date,
        end: Date
    },
    metrics: {
        calories: number
    }
}