export function apiFetch(
  input: Partial<RequestInfo>,
  init?: RequestInit,
  fetchFn = fetch
): Promise<Response> {
  if (typeof input === "string") {
    return fetchFn(
      input.startsWith("/api") ? input : "/api" + input,
      init
        ? {
            ...init,
            headers: {
              ...init.headers,
              "Content-Type":
                init.headers && init.headers["Content-Type"]
                  ? init.headers["Content-Type"]
                  : "application/json",
            },
          }
        : undefined
    );
  }
  return fetchFn(
    {
      ...input,
      url: input.url.startsWith("/api") ? input.url : "/api" + input.url,
    } as RequestInfo,
    init
      ? {
          ...init,
          headers: {
            ...init.headers,
            "Content-Type":
              init.headers && init.headers["Content-Type"]
                ? init.headers["Content-Type"]
                : "application/json",
          },
        }
      : undefined
  );
}

export const DEFAULT_METRICS = {
  calories: { label: "Calories", priority: 0 },
};

// If we ever change this value, make it config driven from API
export const METRIC_MAX = 999999;
