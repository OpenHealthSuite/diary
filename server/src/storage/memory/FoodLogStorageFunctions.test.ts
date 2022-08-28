import { editFoodLog, storeFoodLog } from "./FoodLogStorageFunctions"
import crypto from 'node:crypto'
import { CreateFoodLogEntry, EditFoodLogEntry, isValidationError } from "../types"

describe("FoodLogStorageFunctions", () => {
    describe("CreateFoodLog", () => {
        describe("Validation Errors", () => {

            const GoldInput : CreateFoodLogEntry = {
                name: "My Food Log",
                labels: new Set(["Some Label", "Some other label"]),
                time: {
                    start: new Date(),
                    end: new Date()
                },
                metrics: {
                    calories: 500
                }
            }
    
            const testUserId = crypto.randomUUID()
    
            const {name, ...nameless} = structuredClone(GoldInput);
            const {labels, ...labelless} = structuredClone(GoldInput);
            const {metrics, ...metricless} = structuredClone(GoldInput);
            let weirdMetric: any = structuredClone(GoldInput);
            weirdMetric.metrics.calories = "This is not a number";
            const {time, ...timeless} = structuredClone(GoldInput);
            let startTimeLess: any = structuredClone(GoldInput);
            delete startTimeLess.time.start;
            let endTimeLess: any = structuredClone(GoldInput);
            delete endTimeLess.time.end;
    
            const BadValues: any[] = [
                ["Empty", {}],
                ["WithId", { ...GoldInput, id: crypto.randomUUID() }],
                ["No Name", nameless],
                ["No Labels", labelless],
                ["No Metrics", metricless],
                ["Non-number Metric", weirdMetric],
                ["No Times", timeless],
                ["No Start Time", startTimeLess],
                ["No End Time", endTimeLess]
            ]
    
            it.each(BadValues)("Rejects Bad Test Case %s", async (name: string, badValue: any) => {
                const result = await storeFoodLog(testUserId, badValue)
    
                expect(result.isErr()).toBeTruthy()
                expect(isValidationError(result._unsafeUnwrapErr())).toBeTruthy()
            })
        })
    })

    describe("EditFoodLog", () => {
        describe("Validation Errors", () => {

            const GoldInput : EditFoodLogEntry = {
                id: crypto.randomUUID(),
                name: "My Food Log",
                labels: new Set(["Some Label", "Some other label"]),
                time: {
                    start: new Date(),
                    end: new Date()
                },
                metrics: {
                    calories: 500
                }
            }
    
            const testUserId = crypto.randomUUID()
    
            let weirdMetric: any = structuredClone(GoldInput);
            weirdMetric.metrics.calories = "This is not a number";
            let startTimeLess: any = structuredClone(GoldInput);
            delete startTimeLess.time.start;
            let endTimeLess: any = structuredClone(GoldInput);
            delete endTimeLess.time.end;
    
            const BadValues: any[] = [
                ["Empty", {}],
                ["WithoutId", { ...GoldInput, id: undefined }],
                ["Non-number Metric", weirdMetric],
                ["No Start Time", startTimeLess],
                ["No End Time", endTimeLess]
            ]
    
            it.each(BadValues)("Rejects Bad Test Case %s", async (name: string, badValue: any) => {
                const result = await editFoodLog(testUserId, badValue)
    
                expect(result.isErr()).toBeTruthy()
                expect(isValidationError(result._unsafeUnwrapErr())).toBeTruthy()
            })
        })
    })
})