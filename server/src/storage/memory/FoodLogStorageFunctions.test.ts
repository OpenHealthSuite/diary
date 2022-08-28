import { storeFoodLog } from "./FoodLogStorageFunctions"
import crypto from 'node:crypto'
import { CreateFoodLogEntry, isValidationError } from "../types"

describe("FoodLogStorageFunctions", () => {
    describe("Validation Errors", () => {
        const testUserId = crypto.randomUUID()
        const BadValues: any[] = [
            {}
        ]
        it.each(BadValues)("Rejects Bad Test Case", async (badValue: any) => {
            const result = await storeFoodLog(testUserId, badValue)

            expect(result.isErr()).toBeTruthy()
            expect(isValidationError(result._unsafeUnwrapErr())).toBeTruthy()
        })
    })
})