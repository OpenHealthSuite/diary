import {
  createFoodLogHandler,
  getFoodLogHandler,
  updateFoodLogHandler,
  deleteFoodLogHandler,
  buildRouter,
  queryFoodLogHandler,
  purgeFoodLogHandler,
  exportLogsHandler
} from "./FoodLogHandlers";
import { Request, Response, Router } from "express";
import crypto from "node:crypto";
import { OFDLocals } from "../middlewares";
import { NotFoundError, ValidationError } from "../storage";
import { err, ok } from "neverthrow";
import { CreateFoodLogEntry } from "../storage/types/FoodLog";
import { FoodLogEntry } from "../types";

describe("Handler Registration", () => {
  it("Registers all functions on routes", () => {
    const fakeRouter = {
      route: jest.fn().mockReturnThis(),
      post: jest.fn().mockReturnThis(),
      get: jest.fn().mockReturnThis(),
      put: jest.fn().mockReturnThis(),
      delete: jest.fn().mockReturnThis()
    } as unknown as Router;

    buildRouter(fakeRouter);

    expect(fakeRouter.post).toBeCalledWith("/logs", createFoodLogHandler);
    expect(fakeRouter.get).toBeCalledWith("/logs", queryFoodLogHandler);
    expect(fakeRouter.get).toBeCalledWith("/logs/:itemId", getFoodLogHandler);
    expect(fakeRouter.put).toBeCalledWith(
      "/logs/:itemId",
      updateFoodLogHandler
    );
    expect(fakeRouter.delete).toBeCalledWith(
      "/logs/:itemId",
      deleteFoodLogHandler
    );
    expect(fakeRouter.delete).toBeCalledWith("/logs", purgeFoodLogHandler);
  });
});

describe("Create Food Log Handler", () => {
  test("Happy Path :: Passes to storage, success, returns id of log", async () => {
    const createdId = crypto.randomUUID();
    const mockStorage = jest.fn().mockResolvedValue(ok(createdId));
    const userId = crypto.randomUUID();
    const input: CreateFoodLogEntry = {
      name: "My Log",
      labels: [],
      time: {
        start: new Date(),
        end: new Date()
      },
      metrics: {}
    };

    const fakeReq = {
      body: input
    };

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId
      }
    };

    await createFoodLogHandler(
      fakeReq as Request,
      fakeRes as any as Response & { locals: OFDLocals },
      jest.fn(),
      mockStorage
    );

    expect(mockStorage).toBeCalledTimes(1);
    expect(JSON.stringify(mockStorage.mock.calls[0])).toBe(
      JSON.stringify([userId, input])
    );
    expect(fakeRes.send).toBeCalledWith(createdId);
  });

  test("Validation Error :: Returns error from storage with 400", async () => {
    const validationProblem = "Some Validation Problem";
    const mockStorage = jest
      .fn()
      .mockResolvedValue(err(new ValidationError(validationProblem)));
    const userId = crypto.randomUUID();
    const input: CreateFoodLogEntry = {
      name: "My Log",
      labels: [],
      time: {
        start: new Date(),
        end: new Date()
      },
      metrics: {}
    };

    const fakeReq = {
      body: input
    };

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId
      }
    };

    await createFoodLogHandler(
      fakeReq as Request,
      fakeRes as unknown as Response & { locals: OFDLocals },
      jest.fn(),
      mockStorage
    );

    expect(mockStorage).toBeCalledTimes(1);
    expect(JSON.stringify(mockStorage.mock.calls[0])).toBe(
      JSON.stringify([userId, input])
    );
    expect(fakeRes.status).toBeCalledWith(400);
    expect(fakeRes.send).toBeCalledWith(validationProblem);
  });

  test("Generic Error :: Returns error from storage with 500", async () => {
    const errorMessage = "Some Serious Problem";
    const mockStorage = jest
      .fn()
      .mockResolvedValue(err(new Error(errorMessage)));
    const userId = crypto.randomUUID();
    const input: CreateFoodLogEntry = {
      name: "My Log",
      labels: [],
      time: {
        start: new Date(),
        end: new Date()
      },
      metrics: {}
    };

    const fakeReq = {
      body: input
    };

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId
      }
    };

    await createFoodLogHandler(
      fakeReq as Request,
      fakeRes as unknown as Response & { locals: OFDLocals },
      jest.fn(),
      mockStorage
    );

    expect(mockStorage).toBeCalledTimes(1);
    expect(JSON.stringify(mockStorage.mock.calls[0])).toBe(
      JSON.stringify([userId, input])
    );
    expect(fakeRes.status).toBeCalledWith(500);
    expect(fakeRes.send).toBeCalledWith(errorMessage);
  });
});

describe("Query Food Log Handler", () => {
  test("Happy Path :: Passes to storage, success, returns logs", async () => {
    const userId = crypto.randomUUID();
    const startDate = new Date(1999, 10, 1);
    const endDate = new Date(1999, 11, 1);
    const foodLog: FoodLogEntry[] = [
      {
        id: crypto.randomUUID(),
        name: "My Log",
        labels: [],
        time: {
          start: new Date(),
          end: new Date()
        },
        metrics: {}
      }
    ];

    const mockStorage = jest.fn().mockResolvedValue(ok(foodLog));

    const fakeReq: any = {
      query: {
        startDate: startDate.toISOString(),
        endDate: endDate.toISOString()
      }
    };

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId
      }
    };

    await queryFoodLogHandler(
      fakeReq as Request,
      fakeRes as any as Response & { locals: OFDLocals },
      jest.fn(),
      mockStorage
    );

    expect(mockStorage).toBeCalledTimes(1);
    expect(mockStorage).toBeCalledWith(userId, startDate, endDate);
    expect(fakeRes.send).toBeCalledWith(foodLog);
  });

  const BadDateValues: (string | undefined)[][] = [
    [undefined, undefined],
    [new Date(1987, 10, 1).toISOString(), undefined],
    [undefined, new Date(1987, 10, 1).toISOString()],
    [new Date(1987, 10, 1).toISOString(), "not a date"],
    ["not a date", new Date(1987, 10, 1).toISOString()],
    ["not a date", "also not a date"]
  ];

  test.each(BadDateValues)(
    "Validation :: not valid dates, rejects without trying storage",
    async (startDate, endDate) => {
      const userId = crypto.randomUUID();

      const fakeReq: any = {
        query: {
          startDate,
          endDate
        }
      };

      const mockStorage = jest.fn();

      const fakeRes = {
        send: jest.fn(),
        status: jest.fn().mockReturnThis(),
        locals: {
          userId
        }
      };

      await queryFoodLogHandler(
        fakeReq as Request,
        fakeRes as any as Response & { locals: OFDLocals },
        jest.fn(),
        mockStorage
      );

      expect(mockStorage).not.toBeCalled();
      expect(fakeRes.status).toBeCalledWith(400);
      expect(fakeRes.send).toBeCalledWith("Invalid startDate or endDate");
    }
  );
});

describe("Get Food Log Handler", () => {
  test("Happy Path :: Passes to storage, success, returns log", async () => {
    const userId = crypto.randomUUID();
    const itemId = crypto.randomUUID();
    const foodLog: FoodLogEntry = {
      id: itemId,
      name: "My Log",
      labels: [],
      time: {
        start: new Date(),
        end: new Date()
      },
      metrics: {}
    };

    const mockStorage = jest.fn().mockResolvedValue(ok(foodLog));

    const fakeReq: any = {
      params: {
        itemId
      }
    };

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId
      }
    };

    await getFoodLogHandler(
      fakeReq as Request,
      fakeRes as any as Response & { locals: OFDLocals },
      jest.fn(),
      mockStorage
    );

    expect(mockStorage).toBeCalledTimes(1);
    expect(mockStorage).toBeCalledWith(userId, itemId);
    expect(fakeRes.send).toBeCalledWith(foodLog);
  });

  test("Not found :: returns 404 and message", async () => {
    const userId = crypto.randomUUID();
    const itemId = crypto.randomUUID();

    const errorMessage = "Food Log Not found";

    const mockStorage = jest
      .fn()
      .mockResolvedValue(err(new NotFoundError(errorMessage)));

    const fakeReq: any = {
      params: {
        itemId
      }
    };

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId
      }
    };

    await getFoodLogHandler(
      fakeReq as Request,
      fakeRes as any as Response & { locals: OFDLocals },
      jest.fn(),
      mockStorage
    );

    expect(mockStorage).toBeCalledTimes(1);
    expect(mockStorage).toBeCalledWith(userId, itemId);
    expect(fakeRes.status).toBeCalledWith(404);
    expect(fakeRes.send).toBeCalledWith(errorMessage);
  });
});

describe("Update Food Log Handler", () => {
  test("Happy Path :: Passes to storage, success, returns id of log", async () => {
    const itemId = crypto.randomUUID();
    const userId = crypto.randomUUID();
    const input = {
      name: "My Log",
      labels: new Set<string>(),
      time: {
        start: new Date(),
        end: new Date()
      },
      metrics: {}
    };

    const expectedStorageInput = { id: itemId, ...input };

    const storageResponse = { whoami: "storageResponse" };

    const mockStorage = jest.fn().mockResolvedValue(ok(storageResponse));

    const fakeReq: any = {
      params: {
        itemId
      },
      body: input
    };

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId
      }
    };

    await updateFoodLogHandler(
      fakeReq as Request,
      fakeRes as any as Response & { locals: OFDLocals },
      jest.fn(),
      mockStorage
    );

    expect(mockStorage).toBeCalledTimes(1);
    expect(JSON.stringify(mockStorage.mock.calls[0])).toBe(
      JSON.stringify([userId, expectedStorageInput])
    );
    expect(fakeRes.send.mock.calls[0][0]).toEqual(storageResponse);
  });

  test("Validation Error :: returns 400 with message", async () => {
    const itemId = crypto.randomUUID();
    const userId = crypto.randomUUID();
    const input = {
      name: "My Log",
      labels: new Set<string>()
    };

    const expectedStorageInput = { id: itemId, ...input };

    const storageResponse = new ValidationError("Error validating");

    const mockStorage = jest.fn().mockResolvedValue(err(storageResponse));

    const fakeReq: any = {
      params: {
        itemId
      },
      body: input
    };

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId
      }
    };

    await updateFoodLogHandler(
      fakeReq as Request,
      fakeRes as any as Response & { locals: OFDLocals },
      jest.fn(),
      mockStorage
    );

    expect(mockStorage).toBeCalledTimes(1);
    expect(JSON.stringify(mockStorage.mock.calls[0])).toBe(
      JSON.stringify([userId, expectedStorageInput])
    );
    expect(fakeRes.status).toBeCalledWith(400);
    expect(fakeRes.send).toBeCalledWith(storageResponse.message);
  });

  test("Not Found Error :: returns 404 with message", async () => {
    const itemId = crypto.randomUUID();
    const userId = crypto.randomUUID();
    const input = {
      name: "My Log",
      labels: new Set<string>()
    };

    const expectedStorageInput = { id: itemId, ...input };

    const storageResponse = new NotFoundError("Error finding");

    const mockStorage = jest.fn().mockResolvedValue(err(storageResponse));

    const fakeReq: any = {
      params: {
        itemId
      },
      body: input
    };

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId
      }
    };

    await updateFoodLogHandler(
      fakeReq as Request,
      fakeRes as any as Response & { locals: OFDLocals },
      jest.fn(),
      mockStorage
    );

    expect(mockStorage).toBeCalledTimes(1);
    expect(JSON.stringify(mockStorage.mock.calls[0])).toBe(
      JSON.stringify([userId, expectedStorageInput])
    );
    expect(fakeRes.status).toBeCalledWith(404);
    expect(fakeRes.send).toBeCalledWith(storageResponse.message);
  });

  test("Generic Error :: returns 500 with message", async () => {
    const itemId = crypto.randomUUID();
    const userId = crypto.randomUUID();
    const input = {
      name: "My Log",
      labels: new Set<string>()
    };

    const expectedStorageInput = { id: itemId, ...input };

    const storageResponse = new Error("Error something bad");

    const mockStorage = jest.fn().mockResolvedValue(err(storageResponse));

    const fakeReq: any = {
      params: {
        itemId
      },
      body: input
    };

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId
      }
    };

    await updateFoodLogHandler(
      fakeReq as Request,
      fakeRes as any as Response & { locals: OFDLocals },
      jest.fn(),
      mockStorage
    );

    expect(mockStorage).toBeCalledTimes(1);
    expect(JSON.stringify(mockStorage.mock.calls[0])).toBe(
      JSON.stringify([userId, expectedStorageInput])
    );
    expect(fakeRes.status).toBeCalledWith(500);
    expect(fakeRes.send).toBeCalledWith(storageResponse.message);
  });
});

describe("Get Food Log Handler", () => {
  test("Happy Path :: Passes to storage, 204 success", async () => {
    const userId = crypto.randomUUID();
    const itemId = crypto.randomUUID();

    const mockStorage = jest.fn().mockResolvedValue(ok(true));

    const fakeReq: any = {
      params: {
        itemId
      }
    };

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId
      }
    };

    await deleteFoodLogHandler(
      fakeReq as Request,
      fakeRes as any as Response & { locals: OFDLocals },
      jest.fn(),
      mockStorage
    );

    expect(mockStorage).toBeCalledTimes(1);
    expect(mockStorage).toBeCalledWith(userId, itemId);
    expect(fakeRes.status).toBeCalledWith(204);
    expect(fakeRes.send).toBeCalled();
  });

  test("Generic Error :: returns 500 with message", async () => {
    const userId = crypto.randomUUID();
    const itemId = crypto.randomUUID();
    const storageResponse = new Error("Error something bad");

    const mockStorage = jest.fn().mockResolvedValue(err(storageResponse));

    const fakeReq: any = {
      params: {
        itemId
      }
    };

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId
      }
    };

    await deleteFoodLogHandler(
      fakeReq as Request,
      fakeRes as any as Response & { locals: OFDLocals },
      jest.fn(),
      mockStorage
    );

    expect(mockStorage).toBeCalledTimes(1);
    expect(mockStorage).toBeCalledWith(userId, itemId);
    expect(fakeRes.status).toBeCalledWith(500);
    expect(fakeRes.send).toBeCalledWith(storageResponse.message);
  });
});

describe("Purge  Log Handler", () => {
  test("Happy Path :: Passes to storage, success", async () => {
    const userId = crypto.randomUUID();

    const mockStorage = jest.fn().mockResolvedValue(ok(true));

    const fakeReq: any = {};

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId
      }
    };

    await purgeFoodLogHandler(
      fakeReq as Request,
      fakeRes as any as Response & { locals: OFDLocals },
      jest.fn(),
      mockStorage
    );

    expect(mockStorage).toBeCalledTimes(1);
    expect(mockStorage).toBeCalledWith(userId);
    expect(fakeRes.send).toBeCalled();
  });
  test("Generic Error :: returns 500 with message", async () => {
    const userId = crypto.randomUUID();
    const storageResponse = new Error("Error something bad");

    const mockStorage = jest.fn().mockResolvedValue(err(storageResponse));

    const fakeReq: any = {};

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId
      }
    };

    await purgeFoodLogHandler(
      fakeReq as Request,
      fakeRes as any as Response & { locals: OFDLocals },
      jest.fn(),
      mockStorage
    );

    expect(mockStorage).toBeCalledTimes(1);
    expect(fakeRes.status).toBeCalledWith(500);
    expect(fakeRes.send).toBeCalledWith(storageResponse.message);
  });
});

describe("Export Log Handler", () => {
  test("Happy Path :: Returns temp file with download", async () => {
    const userId = crypto.randomUUID();

    const testPath = "/tmp/my_test_file.csv";

    const mockStorage = jest.fn().mockResolvedValue(ok(testPath));

    const fakeReq: any = {};

    const fakeRes = {
      send: jest.fn(),
      download: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId
      }
    };

    await exportLogsHandler(
      fakeReq as Request,
      fakeRes as any as Response & { locals: OFDLocals },
      jest.fn(),
      mockStorage
    );

    expect(mockStorage).toBeCalledTimes(1);
    expect(mockStorage).toBeCalledWith(userId);
    expect(fakeRes.download).toBeCalledWith(testPath);
  });
  test("Generic Error :: returns 500 with message", async () => {
    const userId = crypto.randomUUID();
    const storageResponse = new Error("Error something bad");

    const mockStorage = jest.fn().mockResolvedValue(err(storageResponse));

    const fakeReq: any = {};

    const fakeRes = {
      send: jest.fn(),
      sendStatus: jest.fn().mockReturnThis(),
      locals: {
        userId
      }
    };

    await exportLogsHandler(
      fakeReq as Request,
      fakeRes as any as Response & { locals: OFDLocals },
      jest.fn(),
      mockStorage
    );

    expect(mockStorage).toBeCalledTimes(1);
    expect(fakeRes.sendStatus).toBeCalledWith(500);
  });
});
