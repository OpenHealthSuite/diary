import { deleteFoodLog, editFoodLog, retrieveFoodLog, storeFoodLog } from "./FoodLogStorageFunctions"
import crypto from 'node:crypto'
import { CreateFoodLogEntry, isNotFoundError, isValidationError } from "../types"


describe("Food Log Storage Integration Tests", () => {
    test("Happy Path :: Bad Retreives, Creates, Retreives, Edits, Reretrieves, Deletes, Fails Retreive, Redelete succeeds false", async () => {
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

        storedItem.name = "Modified Food Log"
        storedItem.metrics = {
            calories: 400
        }

        const modifiedResult = await editFoodLog(testUserId, storedItem)
        expect(modifiedResult.isOk()).toBeTruthy()
        const modified = modifiedResult._unsafeUnwrap()
        expect(modified).toEqual(storedItem)

        const reretrievedItemResult = await retrieveFoodLog(testUserId, testItemId)

        expect(reretrievedItemResult.isOk()).toBeTruthy()
        const reretreived = reretrievedItemResult._unsafeUnwrap()
        expect(reretreived).toEqual(storedItem)

        const deleteResult = await deleteFoodLog(testUserId, testItemId)

        expect(deleteResult.isOk()).toBeTruthy()
        expect(deleteResult._unsafeUnwrap()).toBeTruthy()

        const postDeleteRetrieve = await retrieveFoodLog(testUserId, testItemId)

        expect(postDeleteRetrieve.isErr()).toBeTruthy()
        expect(isNotFoundError(postDeleteRetrieve._unsafeUnwrapErr())).toBeTruthy();

        const redeleteResult = await deleteFoodLog(testUserId, testItemId)

        expect(redeleteResult.isOk()).toBeTruthy()
        expect(redeleteResult._unsafeUnwrap()).toBeFalsy()
    })
})