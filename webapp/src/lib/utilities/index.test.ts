import { apiFetch } from './index'
import { vi } from 'vitest';


describe('apiFetch', () => {
    const inits = [{whoami: "init"}, undefined]
    const inputUrls = ["/someroute?this=query", "/route", "/this/is/nested"]
    const testMatrix: any[][] = inputUrls.reduce((prev, curr) => [...prev, ...inits.map(m => [curr, m])], [])

    const alreadyApi = ["/api/somewhere", { url: "/api/somehow"}, "/api", { url: "/api"}]

    it.each(alreadyApi)('Given already has /api at start :: does not prepend', async (input) => {
        const response = { whoami: "fetchResponse" };
        const fakeFetch = vi.fn().mockResolvedValueOnce(response);

        const result = await apiFetch(input, undefined, fakeFetch);

        expect(fakeFetch).toBeCalledWith(input, undefined)
        expect(result).toBe(response)
    })

    it.each(testMatrix)('Given string :: prepends "/api" and drops it into fetch', async (url, init) => {
        const response = { whoami: "fetchResponse" };
        const fakeFetch = vi.fn().mockResolvedValueOnce(response);

        const result = await apiFetch(url, init as any, fakeFetch);

        expect(fakeFetch).toBeCalledWith("/api" + url, init)
        expect(result).toBe(response)
    })

    it.each(testMatrix)('Given request object :: prepends "/api" and drops it into fetch', async (url, init) => {
        const response = { whoami: "fetchResponse" };
        const fakeFetch = vi.fn().mockResolvedValueOnce(response);

        const result = await apiFetch({ url }, init as any, fakeFetch);

        expect(fakeFetch).toBeCalledWith({ url: '/api' + url }, init)
        expect(result).toBe(response)
    })

    it.each(testMatrix)('Given request object with other params :: prepends "/api" and drops it into fetch', async (url, init) => {
        const response = { whoami: "fetchResponse" };
        const fakeFetch = vi.fn().mockResolvedValueOnce(response);

        const input = { url, integrity: "someValue" };
        const result = await apiFetch(input, init as any, fakeFetch);

        expect(fakeFetch).toBeCalledWith({ url: '/api' + url, integrity: "someValue" }, init)
        expect(result).toBe(response)
    })
})
