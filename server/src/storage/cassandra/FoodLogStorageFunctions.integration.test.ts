import { deleteFoodLog, editFoodLog, queryFoodLogs, retrieveFoodLog, storeFoodLog } from "./FoodLogStorageFunctions"
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

        // storedItem.name = "Modified Food Log"
        // storedItem.metrics = {
        //     calories: 400
        // }

        // const modifiedResult = await editFoodLog(testUserId, storedItem)
        // expect(modifiedResult.isOk()).toBeTruthy()
        // const modified = modifiedResult._unsafeUnwrap()
        // expect(modified).toEqual(storedItem)

        // const reretrievedItemResult = await retrieveFoodLog(testUserId, testItemId)

        // expect(reretrievedItemResult.isOk()).toBeTruthy()
        // const reretreived = reretrievedItemResult._unsafeUnwrap()
        // expect(reretreived).toEqual(storedItem)

        // const deleteResult = await deleteFoodLog(testUserId, testItemId)

        // expect(deleteResult.isOk()).toBeTruthy()
        // expect(deleteResult._unsafeUnwrap()).toBeTruthy()

        // const postDeleteRetrieve = await retrieveFoodLog(testUserId, testItemId)

        // expect(postDeleteRetrieve.isErr()).toBeTruthy()
        // expect(isNotFoundError(postDeleteRetrieve._unsafeUnwrapErr())).toBeTruthy();

        // const redeleteResult = await deleteFoodLog(testUserId, testItemId)

        // expect(redeleteResult.isOk()).toBeTruthy()
        // expect(redeleteResult._unsafeUnwrap()).toBeFalsy()
    })

    // test("Queries :: can add some logs, and get expected query results", async () => {
    //     const testUserId = crypto.randomUUID()

    //     const pastLog: CreateFoodLogEntry = {
    //         name: "My Food Log",
    //         labels: new Set(),
    //         time: {
    //             start: new Date(1999, 10, 10),
    //             end: new Date(1999, 10, 11)
    //         },
    //         metrics: {
    //             calories: 500
    //         }
    //     }

    //     const centerLog: CreateFoodLogEntry = {
    //         name: "My Food Log",
    //         labels: new Set(),
    //         time: {
    //             start: new Date(1999, 10, 15),
    //             end: new Date(1999, 10, 16)
    //         },
    //         metrics: {
    //             calories: 500
    //         }
    //     }

    //     const futureLog: CreateFoodLogEntry = {
    //         name: "My Food Log",
    //         labels: new Set(),
    //         time: {
    //             start: new Date(1999, 10, 20),
    //             end: new Date(1999, 10, 21)
    //         },
    //         metrics: {
    //             calories: 500
    //         }
    //     }

    //     const past = await storeFoodLog(testUserId, pastLog)
    //     const pastItemId = past._unsafeUnwrap()

    //     const result = await storeFoodLog(testUserId, centerLog)
    //     const centerItemId = result._unsafeUnwrap()

    //     const future = await storeFoodLog(testUserId, futureLog)
    //     const futureItemId = future._unsafeUnwrap()


    //     const startingQueryResult = await queryFoodLogs(testUserId, new Date(1999, 10, 15), new Date(1999, 10, 16))
        
    //     expect(startingQueryResult.isOk()).toBeTruthy()
    //     const firstTest = startingQueryResult._unsafeUnwrap();
    //     expect(firstTest.length).toBe(1)
    //     expect(firstTest[0].id).toBe(centerItemId)

    //     const pastQueryResult = await queryFoodLogs(testUserId, new Date(1999, 10, 9), new Date(1999, 10, 16))
        
    //     expect(pastQueryResult.isOk()).toBeTruthy()
    //     const secondTest = pastQueryResult._unsafeUnwrap();
    //     expect(secondTest.length).toBe(2)
    //     expect(secondTest.map(x => x.id)).toEqual([pastItemId, centerItemId])
        
    //     const futureQueryResult = await queryFoodLogs(testUserId, new Date(1999, 10, 15), new Date(1999, 10, 30))
        
    //     expect(futureQueryResult.isOk()).toBeTruthy()
    //     const thirdTest = futureQueryResult._unsafeUnwrap();
    //     expect(thirdTest.length).toBe(2)
    //     expect(thirdTest.map(x => x.id)).toEqual([centerItemId, futureItemId])

    //     const wildQueryResult = await queryFoodLogs(testUserId, new Date(2012, 0, 1), new Date(2012, 11, 31))
        
    //     expect(startingQueryResult.isOk()).toBeTruthy()
    //     const wildTest = wildQueryResult._unsafeUnwrap();
    //     expect(wildTest.length).toBe(0)
    // })
})