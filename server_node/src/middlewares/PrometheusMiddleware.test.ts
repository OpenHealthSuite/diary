import { promMiddleware } from "./PrometheusMiddleware";

describe("promMiddleware", () => {
  const requests = [
    { path: "/somewhere", type: "client", method: "GET", status: 200 },
    { path: "/api/someendpoint", type: "api", method: "POST", status: 200 },
    { path: "/api", type: "api", method: "PUT", status: 204 },
    { path: "/", type: "client", method: "DELETE", status: 204 },
    { path: "/settings", type: "client", method: "GET", status: 500 }
  ];
  test.each(requests)(
    "Increments request counter",
    ({ path, type, method, status }) => {
      const fakeCounter = {
        labels: jest.fn().mockReturnThis(),
        inc: jest.fn()
      };
      const fakeNext = jest.fn();

      promMiddleware(
        { method, path } as any,
        { statusCode: status } as any,
        fakeNext,
        fakeCounter as any
      );

      expect(fakeNext).toBeCalled();
      expect(fakeCounter.labels).toBeCalledWith(
        type,
        method,
        status.toString()
      );
      expect(fakeCounter.inc).toBeCalledTimes(1);
    }
  );

  test("Ignores /metrics path", () => {
    const fakeCounter = {
      labels: jest.fn().mockReturnThis(),
      inc: jest.fn()
    };
    const fakeNext = jest.fn();

    promMiddleware(
      { method: "GET", path: "/metrics" } as any,
      { statusCode: 200 } as any,
      fakeNext,
      fakeCounter as any
    );

    expect(fakeNext).toBeCalled();
    expect(fakeCounter.labels).not.toBeCalled();
    expect(fakeCounter.inc).not.toBeCalled();
  });
});
