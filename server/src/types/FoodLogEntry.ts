export type FoodLogEntry = {
    id: string,
    name: string,
    labels: string[],
    time: {
        start: Date,
        end: Date
    },
    metrics: {
        [key: string]: number
    }
}
