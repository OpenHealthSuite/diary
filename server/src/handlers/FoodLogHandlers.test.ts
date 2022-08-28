import { createFoodLogHandler, getFoodLogHandler, updateFoodLogHandler, deleteFoodLogHandler, buildRouter } from './FoodLogHandlers'
import { Express, Request, Response, Router } from 'express';
import crypto from 'node:crypto';
import { OFDLocals } from '../middlewares';
import { NotFoundError, ValidationError } from '../storage';
import { err, ok } from 'neverthrow';
import { CreateFoodLogEntry, EditFoodLogEntry } from '../storage/types/FoodLog';
import { FoodLogEntry } from '../types';

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

    expect(fakeRouter.post).toBeCalledWith('/logs', createFoodLogHandler)
    expect(fakeRouter.get).toBeCalledWith('/logs/:logId', getFoodLogHandler)
    expect(fakeRouter.put).toBeCalledWith('/logs/:logId', updateFoodLogHandler)
    expect(fakeRouter.delete).toBeCalledWith('/logs/:logId', deleteFoodLogHandler)
  })
})

describe("Create Food Log Handler", () => {
  test("Happy Path :: Passes to storage, success, returns id of log", async () => {
    const createdId = crypto.randomUUID();
    const mockStorage = jest.fn().mockResolvedValue(ok(createdId));
    const userId = crypto.randomUUID();
    const input: CreateFoodLogEntry = {
      name: 'My Log',
      labels: new Set<string>(),
      time: {
        start: new Date(),
        end: new Date()
      },
      metrics: {}
    }

    const fakeReq = {
      body: input,
    }

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId: userId
      }
    }

    await createFoodLogHandler(fakeReq as Request, fakeRes as any as Response & { locals: OFDLocals }, jest.fn(), mockStorage)

    expect(mockStorage).toBeCalledTimes(1)
    expect(mockStorage).toBeCalledWith(userId, input)
    expect(fakeRes.send).toBeCalledWith(createdId)
  })

  test("Validation Error :: Returns error from storage with 400", async () => {
    const validationProblem = "Some Validation Problem";
    const mockStorage = jest.fn().mockResolvedValue(err(new ValidationError(validationProblem)));
    const userId = crypto.randomUUID();
    const input: CreateFoodLogEntry = {
      name: 'My Log',
      labels: new Set<string>(),
      time: {
        start: new Date(),
        end: new Date()
      },
      metrics: {}
    }

    const fakeReq = {
      body: input,
    }

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId: userId
      }
    }

    await createFoodLogHandler(fakeReq as Request, fakeRes as unknown as Response & { locals: OFDLocals }, jest.fn(), mockStorage)

    expect(mockStorage).toBeCalledTimes(1)
    expect(mockStorage).toBeCalledWith(userId, input)
    expect(fakeRes.status).toBeCalledWith(400)
    expect(fakeRes.send).toBeCalledWith(validationProblem)
  })

  test("Generic Error :: Returns error from storage with 500", async () => {
    const errorMessage = "Some Serious Problem";
    const mockStorage = jest.fn().mockResolvedValue(err(new Error(errorMessage)));
    const userId = crypto.randomUUID();
    const input: CreateFoodLogEntry = {
      name: 'My Log',
      labels: new Set<string>(),
      time: {
        start: new Date(),
        end: new Date()
      },
      metrics: {}
    }

    const fakeReq = {
      body: input,
    }

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId: userId
      }
    }

    await createFoodLogHandler(fakeReq as Request, fakeRes as unknown as Response & { locals: OFDLocals }, jest.fn(), mockStorage)

    expect(mockStorage).toBeCalledTimes(1)
    expect(mockStorage).toBeCalledWith(userId, input)
    expect(fakeRes.status).toBeCalledWith(500)
    expect(fakeRes.send).toBeCalledWith(errorMessage)
  })
})

describe("Get Food Log Handler", () => {
  test("Happy Path :: Passes to storage, success, returns log", async () => {
    const userId = crypto.randomUUID();
    const itemId = crypto.randomUUID();
    const foodLog: FoodLogEntry = {
      id: itemId,
      name: 'My Log',
      labels: new Set<string>(),
      time: {
        start: new Date(),
        end: new Date()
      },
      metrics: {}
    }


    const mockStorage = jest.fn().mockResolvedValue(ok(foodLog));

    const fakeReq: any = {
      params: {
        itemId: itemId
      }
    }

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId: userId
      }
    }

    await getFoodLogHandler(fakeReq as Request, fakeRes as any as Response & { locals: OFDLocals }, jest.fn(), mockStorage)

    expect(mockStorage).toBeCalledTimes(1)
    expect(mockStorage).toBeCalledWith(userId, itemId)
    expect(fakeRes.send).toBeCalledWith(foodLog)
  })

  test("Not found :: returns 404 and message", async () => {
    const userId = crypto.randomUUID();
    const itemId = crypto.randomUUID();

    const errorMessage = "Food Log Not found"

    const mockStorage = jest.fn().mockResolvedValue(err(new NotFoundError(errorMessage)));

    const fakeReq: any = {
      params: {
        itemId: itemId
      }
    }

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId: userId
      }
    }

    await getFoodLogHandler(fakeReq as Request, fakeRes as any as Response & { locals: OFDLocals }, jest.fn(), mockStorage)

    expect(mockStorage).toBeCalledTimes(1)
    expect(mockStorage).toBeCalledWith(userId, itemId)
    expect(fakeRes.status).toBeCalledWith(404)
    expect(fakeRes.send).toBeCalledWith(errorMessage)
  })
})


describe("Update Food Log Handler", () => {
  test("Happy Path :: Passes to storage, success, returns id of log", async () => {
    const itemId = crypto.randomUUID();
    const userId = crypto.randomUUID();
    const input: EditFoodLogEntry = {
      name: 'My Log',
      labels: new Set<string>(),
      time: {
        start: new Date(),
        end: new Date()
      },
      metrics: {}
    }

    const expectedStorageInput = { id: itemId, ...input }

    const storageResponse = { whoami: "storageResponse" }

    const mockStorage = jest.fn().mockResolvedValue(ok(storageResponse));

    const fakeReq: any = {
      params: {
        itemId
      },
      body: input,
    }

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId: userId
      }
    }

    await updateFoodLogHandler(fakeReq as Request, fakeRes as any as Response & { locals: OFDLocals }, jest.fn(), mockStorage)

    expect(mockStorage).toBeCalledTimes(1)
    expect(mockStorage).toBeCalledWith(userId, expectedStorageInput)
    expect(fakeRes.send).toBeCalledWith(storageResponse)
  })

  test("Validation Error :: returns 400 with message", async () => {
    const itemId = crypto.randomUUID();
    const userId = crypto.randomUUID();
    const input: EditFoodLogEntry = {
      name: 'My Log',
      labels: new Set<string>()
    }

    const expectedStorageInput = { id: itemId, ...input }

    const storageResponse = new ValidationError("Error validating")

    const mockStorage = jest.fn().mockResolvedValue(err(storageResponse));

    const fakeReq: any = {
      params: {
        itemId
      },
      body: input,
    }

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId: userId
      }
    }

    await updateFoodLogHandler(fakeReq as Request, fakeRes as any as Response & { locals: OFDLocals }, jest.fn(), mockStorage)

    expect(mockStorage).toBeCalledTimes(1)
    expect(mockStorage).toBeCalledWith(userId, expectedStorageInput)
    expect(fakeRes.status).toBeCalledWith(400)
    expect(fakeRes.send).toBeCalledWith(storageResponse.message)
  })

  test("Not Found Error :: returns 404 with message", async () => {
    const itemId = crypto.randomUUID();
    const userId = crypto.randomUUID();
    const input: EditFoodLogEntry = {
      name: 'My Log',
      labels: new Set<string>()
    }

    const expectedStorageInput = { id: itemId, ...input }

    const storageResponse = new NotFoundError("Error finding")

    const mockStorage = jest.fn().mockResolvedValue(err(storageResponse));

    const fakeReq: any = {
      params: {
        itemId
      },
      body: input,
    }

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId: userId
      }
    }

    await updateFoodLogHandler(fakeReq as Request, fakeRes as any as Response & { locals: OFDLocals }, jest.fn(), mockStorage)

    expect(mockStorage).toBeCalledTimes(1)
    expect(mockStorage).toBeCalledWith(userId, expectedStorageInput)
    expect(fakeRes.status).toBeCalledWith(404)
    expect(fakeRes.send).toBeCalledWith(storageResponse.message)
  })

  test("Generic Error :: returns 500 with message", async () => {
    const itemId = crypto.randomUUID();
    const userId = crypto.randomUUID();
    const input: EditFoodLogEntry = {
      name: 'My Log',
      labels: new Set<string>()
    }

    const expectedStorageInput = { id: itemId, ...input }

    const storageResponse = new Error("Error something bad")

    const mockStorage = jest.fn().mockResolvedValue(err(storageResponse));

    const fakeReq: any = {
      params: {
        itemId
      },
      body: input,
    }

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId: userId
      }
    }

    await updateFoodLogHandler(fakeReq as Request, fakeRes as any as Response & { locals: OFDLocals }, jest.fn(), mockStorage)

    expect(mockStorage).toBeCalledTimes(1)
    expect(mockStorage).toBeCalledWith(userId, expectedStorageInput)
    expect(fakeRes.status).toBeCalledWith(500)
    expect(fakeRes.send).toBeCalledWith(storageResponse.message)
  })
})


describe("Get Food Log Handler", () => {
  test("Happy Path :: Passes to storage, 204 success", async () => {
    const userId = crypto.randomUUID();
    const itemId = crypto.randomUUID();

    const mockStorage = jest.fn().mockResolvedValue(ok(true));

    const fakeReq: any = {
      params: {
        itemId: itemId
      }
    }

    const fakeRes = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
        userId: userId
      }
    }

    await deleteFoodLogHandler(fakeReq as Request, fakeRes as any as Response & { locals: OFDLocals }, jest.fn(), mockStorage)

    expect(mockStorage).toBeCalledTimes(1)
    expect(mockStorage).toBeCalledWith(userId, itemId)
    expect(fakeRes.status).toBeCalledWith(204)
    expect(fakeRes.send).toBeCalled()
  })
})