import { TEST_CONFIGURATION } from "../config"
import crypto from 'node:crypto'

describe("Authentication", () => {
    const testPaths = [
        '/api',
        '/api/logs'
    ]
    it.each(testPaths)("No userid header :: returns 403", async (path) => {
        const response = await fetch(TEST_CONFIGURATION.API_HOST + path)

        expect(response.status).toBe(403)
    })
    it.each(testPaths)("UserId header :: returns not 403", async (path) => {
        const response = await fetch(TEST_CONFIGURATION.API_HOST + path, {
            headers: {
                [TEST_CONFIGURATION.USERID_HEADER]: crypto.randomUUID()
            }
        })

        expect(response.status).not.toBe(403)
    })

    it("Static root :: returns 200", async () => {
        const response = await fetch(TEST_CONFIGURATION.API_HOST + '/')

        expect(response.status).toBe(200)
    })
})