import { retrieveFoodLog, storeFoodLog } from "./FoodLogStorageFunctions"
import crypto from 'node:crypto'
import { CreateFoodLogEntry, isNotFoundError, isValidationError } from "../types"


describe("Food Log Storage Integration Tests", () => {
    test("Happy Path :: Bad Retreives, Creates, Retreives, Edits, Reretrieves, Deletes, Fails Retreive", async () => {
        const testUserId = crypto.randomUUID()

        const randomId = crypto.randomUUID()

        const badResult = await retrieveFoodLog(testUserId, randomId)

        expect(badResult.isErr()).toBeTruthy()
        expect(isNotFoundError(badResult._unsafeUnwrapErr())).toBeTruthy();

        const input: CreateFoodLogEntry = {
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

        const result = await storeFoodLog(testUserId, input)

        expect(result.isOk()).toBeTruthy()
        const testItemId = result._unsafeUnwrap()
        expect(testItemId.length).toBeGreaterThan(0);

        const storedItemResult = await retrieveFoodLog(testUserId, testItemId)

        expect(storedItemResult.isOk()).toBeTruthy()
        const storedItem = storedItemResult._unsafeUnwrap()
        expect(storedItem).toEqual({ id: testItemId, ...input })
    })
})