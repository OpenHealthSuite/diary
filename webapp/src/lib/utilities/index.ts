export function apiFetch(input: Partial<RequestInfo>, init?: RequestInit, fetchFn = fetch): Promise<Response> {
    if (typeof input === "string") {
        return fetchFn(input.startsWith('/api') ? input : '/api' + input, init);
    }
    return fetchFn({ ...input, url: input.url.startsWith('/api') ? input.url : '/api' + input.url } as RequestInfo, init);
}