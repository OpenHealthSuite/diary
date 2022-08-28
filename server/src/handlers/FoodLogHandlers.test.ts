import { addHandlers, createFoodLogHandler } from './FoodLogHandlers'
import { Express, Request, Response} from 'express';
import crypto from 'node:crypto';
import { OFDLocals } from '../middlewares';
import { ValidationError } from '../storage';
import { err, ok } from 'neverthrow';
import { CreateFoodLogEntry } from '../storage/types/FoodLog';

describe("Handler Registration", () => {
  it("Registers all functions on routes", () => {
    const fakeApp = {
      post: jest.fn()
    } as unknown as Express;
    addHandlers(fakeApp);

    expect(fakeApp.post).toBeCalledWith('/logs', createFoodLogHandler)
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