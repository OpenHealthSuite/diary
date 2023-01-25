import { promMiddleware } from "./PrometheusMiddleware";

describe("promMiddleware", () => {
  const requests: [string, number][] = [
    ["GET", 200],
    ["POST", 200],
    ["PUT", 204],
    ["DELETE", 204],
    ["GET", 500],
  ];
  test.each(requests)(
    "Increments request counter :: %s %s",
    ([method, status]) => {
      const fakeCounter = {
        labels: jest.fn().mockReturnThis(),
        inc: jest.fn(),
      };
      const fakeNext = jest.fn();

      promMiddleware(
        { method } as any,
        { statusCode: status } as any,
        fakeNext,
        fakeCounter as any
      );

      expect(fakeNext).toBeCalled();
      expect(fakeCounter.labels).toBeCalledWith(method, status.toString());
      expect(fakeCounter.inc).toBeCalledTimes(1);
    }
  );

  test("Ignores /metrics path", () => {
    const fakeCounter = {
      labels: jest.fn().mockReturnThis(),
      inc: jest.fn(),
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
