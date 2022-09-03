import { TEST_CONFIGURATION } from "../config"
import crypto from 'node:crypto'

describe("Food Log Storage Integration Tests", () => {
    
    test("Happy Path :: Bad Retreives, Creates, Retreives, Edits, Reretrieves, Deletes, Fails Retreive, Redelete succeeds false", async () => {
        const testUserId = crypto.randomUUID()

        const randomId = crypto.randomUUID()

        const badResult = await fetch(TEST_CONFIGURATION.API_HOST + '/api/logs/' + randomId, {
            headers: {
                [TEST_CONFIGURATION.USERID_HEADER]: testUserId
            }
        })

        expect(badResult.status).toBe(404)

        const date = new Date();
        const endDate = new Date(date);
        endDate.setMinutes(endDate.getMinutes() + 5)

        const input = {
            name: "My Food Log",
            labels: ["Some Label", "Some other label"],
            time: {
                start: date.toISOString(),
                end: endDate.toISOString()
            },
            metrics: {
                calories: 500
            }
        }

        const result = await fetch(TEST_CONFIGURATION.API_HOST + '/api/logs' , {
            method: 'POST',
            headers: {
                [TEST_CONFIGURATION.USERID_HEADER]: testUserId,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(input)
        })

        expect(result.status).toBe(200)
        const testItemId = await result.text()
        expect(testItemId.length).toBeGreaterThan(0);

        const storedItemResult = await fetch(TEST_CONFIGURATION.API_HOST + '/api/logs/' + testItemId, {
            headers: {
                [TEST_CONFIGURATION.USERID_HEADER]: testUserId
            }
        })

        expect(storedItemResult.status).toBe(200)
        const storedItem = await storedItemResult.json()
        expect(storedItem).toEqual({ id: testItemId, ...input })

        storedItem.name = "Modified Food Log"
        storedItem.metrics = {
            calories: 400
        }

        const modifiedResult = await fetch(TEST_CONFIGURATION.API_HOST + '/api/logs/' + testItemId, {
            method: 'PUT',
            headers: {
                [TEST_CONFIGURATION.USERID_HEADER]: testUserId,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(storedItem)
        })
        const modified = await modifiedResult.json();
        expect(modified).toEqual(storedItem)

        const reRetreivedResult = await fetch(TEST_CONFIGURATION.API_HOST + '/api/logs/' + testItemId, {
            headers: {
                [TEST_CONFIGURATION.USERID_HEADER]: testUserId
            }
        })

        const reretreived = await reRetreivedResult.json()
        expect(storedItem).toEqual(reretreived)

        const deleteResult = await fetch(TEST_CONFIGURATION.API_HOST + '/api/logs/' + testItemId, {
            method: 'DELETE',
            headers: {
                [TEST_CONFIGURATION.USERID_HEADER]: testUserId
            }
        })

        expect(deleteResult.status).toBe(204);

        const postDeleteReretreive = await fetch(TEST_CONFIGURATION.API_HOST + '/api/logs/' + testItemId, {
            headers: {
                [TEST_CONFIGURATION.USERID_HEADER]: testUserId
            }
        })

        expect(postDeleteReretreive.status).toBe(404)

        const redeleteResult = await fetch(TEST_CONFIGURATION.API_HOST + '/api/logs/' + testItemId, {
            method: 'DELETE',
            headers: {
                [TEST_CONFIGURATION.USERID_HEADER]: testUserId
            }
        })

        expect(redeleteResult.status).toBe(204);
    })

    test("Queries :: can add some logs, and get expected query results", async () => {
        const testUserId = crypto.randomUUID()

        const pastLog = {
            name: "My Food Log",
            labels: [],
            time: {
                start: new Date(1999, 10, 10),
                end: new Date(1999, 10, 11)
            },
            metrics: {
                calories: 500
            }
        }

        const centerLog = {
            name: "My Food Log",
            labels: [],
            time: {
                start: new Date(1999, 10, 15),
                end: new Date(1999, 10, 16)
            },
            metrics: {
                calories: 500
            }
        }

        const futureLog = {
            name: "My Food Log",
            labels: [],
            time: {
                start: new Date(1999, 10, 20),
                end: new Date(1999, 10, 21)
            },
            metrics: {
                calories: 500
            }
        }

        const pastItemId = await(await fetch(TEST_CONFIGURATION.API_HOST + '/api/logs' , {
            method: 'POST',
            headers: {
                [TEST_CONFIGURATION.USERID_HEADER]: testUserId,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(pastLog)
        })).text()

        const centerItemId = await(await fetch(TEST_CONFIGURATION.API_HOST + '/api/logs' , {
            method: 'POST',
            headers: {
                [TEST_CONFIGURATION.USERID_HEADER]: testUserId,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(centerLog)
        })).text()

        const futureItemId = await(await fetch(TEST_CONFIGURATION.API_HOST + '/api/logs' , {
            method: 'POST',
            headers: {
                [TEST_CONFIGURATION.USERID_HEADER]: testUserId,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(futureLog)
        })).text()

        const queryLogs = async (startDate: Date, endDate: Date) => {
            return await(await fetch(TEST_CONFIGURATION.API_HOST + `/api/logs?startDate=${startDate.toISOString()}&endDate=${endDate.toISOString()}`, {
                headers: {
                    [TEST_CONFIGURATION.USERID_HEADER]: testUserId
                }
            })).json()
        }

        const firstTest = await queryLogs(new Date(1999, 10, 15), new Date(1999, 10, 16))
        
        expect(firstTest.length).toBe(1)
        expect(firstTest[0].id).toBe(centerItemId)

        const secondTest = await queryLogs(new Date(1999, 10, 9), new Date(1999, 10, 16))
        
        expect(secondTest.length).toBe(2)
        expect(secondTest.map((x: any) => x.id).sort()).toEqual([pastItemId, centerItemId].sort())
        
        const thirdTest = await queryLogs(new Date(1999, 10, 15), new Date(1999, 10, 30))
        
        expect(thirdTest.length).toBe(2)
        expect(thirdTest.map((x: any) => x.id).sort()).toEqual([centerItemId, futureItemId].sort())

        const wildTest = await queryLogs(new Date(2012, 0, 1), new Date(2012, 11, 31))
        
        expect(wildTest.length).toBe(0)
    })
})