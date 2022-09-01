import { apiFetch } from './index'
import { vi } from 'vitest';


describe('apiFetch', () => {
    const inits = [{whoami: "init"}, undefined]
    const inputUrls = ["/someroute?this=query", "/route", "/this/is/nested"]
    const testMatrix: any[][] = inputUrls.reduce((prev, curr) => [...prev, ...inits.map(m => [curr, m])], [])

    const alreadyApiGet = ["/api/somewhere", "/api"]
    const alreadyApiObject = [{ url: "/api/somehow"}, { url: "/api" }]

    it.each(alreadyApiGet)('Given already has /api at start :: does not prepend', async (input) => {
        const response = { whoami: "fetchResponse" };
        const fakeFetch = vi.fn().mockResolvedValueOnce(response);

        const result = await apiFetch(input, undefined, fakeFetch);

        expect(fakeFetch).toBeCalledWith(input, undefined)
        expect(result).toBe(response)
    })    
    
    it.each(alreadyApiObject)('Given already has /api at start object :: does not prepend', async (input) => {
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

        expect(fakeFetch).toBeCalledWith("/api" + url, init ? { ...init, headers: { 'Content-Type': 'application/json' } } : undefined)
        expect(result).toBe(response)
    })

    it.each(testMatrix)('Given request object :: prepends "/api" and drops it into fetch', async (url, init) => {
        const response = { whoami: "fetchResponse" };
        const fakeFetch = vi.fn().mockResolvedValueOnce(response);

        const result = await apiFetch({ url }, init as any, fakeFetch);

        expect(fakeFetch).toBeCalledWith({ url: '/api' + url,  }, init ? { ...init, headers: { 'Content-Type': 'application/json' } } : undefined)
        expect(result).toBe(response)
    })

    it.each(testMatrix)('Given request object with other params :: prepends "/api" and drops it into fetch', async (url, init) => {
        const response = { whoami: "fetchResponse" };
        const fakeFetch = vi.fn().mockResolvedValueOnce(response);

        const input = { url, integrity: "someValue" };
        const result = await apiFetch(input, init as any, fakeFetch);

        expect(fakeFetch).toBeCalledWith({ url: '/api' + url, integrity: "someValue" }, init ? { ...init, headers: { 'Content-Type': 'application/json' } } : undefined)
        expect(result).toBe(response)
    })

    const methodsWithContent = ["POST", "PUT"]

    it.each(methodsWithContent)('Given %s request object without provided content type header or any headers :: sets application/json', async (method) => {
        const response = { whoami: "fetchResponse" };
        const fakeFetch = vi.fn().mockResolvedValueOnce(response);

        const input = { url: "/somewhere" };
        const result = await apiFetch(input, { method } as any, fakeFetch);

        expect(fakeFetch).toBeCalledWith({ url: '/api' + "/somewhere" }, { method, headers: { 'Content-Type': 'application/json' }})
        expect(result).toBe(response)
    })


    const inputCTHeaders = ["text/plain", "image/bmp"]
    it.each(inputCTHeaders)('Given %s request object with provided content type header :: does not overwrite header', async (header) => {
        const response = { whoami: "fetchResponse" };
        const fakeFetch = vi.fn().mockResolvedValueOnce(response);

        const input = { url: "/somewhere" };
        const result = await apiFetch(input as any, { headers: { 'Content-Type': header } } as any, fakeFetch);

        expect(fakeFetch).toBeCalledWith({ url: '/api' + "/somewhere" }, { headers: { 'Content-Type': header }})
        expect(result).toBe(response)
    })

    it("Given request with many headers but not content type :: sets content type and preserves rest", async () => {
        const response = { whoami: "fetchResponse" };
        const fakeFetch = vi.fn().mockResolvedValueOnce(response);

        const input = { url: "/somewhere" };
        const init = { headers: { 'someheader': 'aheaders', 'anotherheader' : 'theotherheader' } } ;
        const result = await apiFetch(input as any, init as any, fakeFetch);

        expect(fakeFetch).toBeCalledWith({ url: '/api' + "/somewhere" }, { headers: { ...init.headers, 'Content-Type': 'application/json' } })
        expect(result).toBe(response)
    })
})
