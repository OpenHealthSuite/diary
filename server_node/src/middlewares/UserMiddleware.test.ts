import { userMiddleware } from "./UserMiddleware";
import crypto from "node:crypto";
import { OFDLocals } from "./types/Locals";

describe("UserMiddleware", () => {
  const INITIAL_ENV = process.env;

  beforeEach(() => {
    process.env = { ...INITIAL_ENV };
  });

  afterAll(() => {
    process.env = INITIAL_ENV;
  });

  test("Happy Path :: By default, takes trusted header and adds to locals", () => {
    const nextfn = jest.fn();
    const testUserId = crypto.randomUUID();
    const req = {
      headers: {
        "x-openfooddiary-userid": testUserId
      }
    };
    const res = {
      locals: {

      }
    };

    userMiddleware(req as any, res as any, nextfn);

    expect(nextfn).toBeCalledTimes(1);
    expect(nextfn).toBeCalledWith();
    expect((res.locals as OFDLocals).userId).toBe(testUserId);
  });

  test("Use different trusted header", () => {
    process.env.OPENFOODDIARY_USERIDHEADER = "x-userid";
    const nextfn = jest.fn();
    const testUserId = crypto.randomUUID();
    const req = {
      headers: {
        [process.env.OPENFOODDIARY_USERIDHEADER]: testUserId
      }
    };
    const res = {
      locals: {

      }
    };

    userMiddleware(req as any, res as any, nextfn);

    expect(nextfn).toBeCalledTimes(1);
    expect(nextfn).toBeCalledWith();
    expect((res.locals as OFDLocals).userId).toBe(testUserId);
  });

  test("Dev ID Set, returns dev ID rather than header id", () => {
    process.env.OPENFOODDIARY_USERID = crypto.randomUUID();
    const nextfn = jest.fn();
    const testUserId = crypto.randomUUID();
    const req = {
      headers: {
        "x-openfooddiary-userid": testUserId
      }
    };
    const res = {
      locals: {

      }
    };

    userMiddleware(req as any, res as any, nextfn);

    expect(nextfn).toBeCalledTimes(1);
    expect(nextfn).toBeCalledWith();
    expect((res.locals as OFDLocals).userId).toBe(process.env.OPENFOODDIARY_USERID);
  });

  test("Dev ID Set and no header, sets dev id", () => {
    process.env.OPENFOODDIARY_USERID = crypto.randomUUID();
    const nextfn = jest.fn();
    const req = {
      headers: {}
    };
    const res = {
      locals: {

      }
    };

    userMiddleware(req as any, res as any, nextfn);

    expect(nextfn).toBeCalledTimes(1);
    expect(nextfn).toBeCalledWith();
    expect((res.locals as OFDLocals).userId).toBe(process.env.OPENFOODDIARY_USERID);
  });

  test("No dev id or header id :: rejects request", () => {
    const nextfn = jest.fn();
    const req = {
      headers: {}
    };
    const res = {
      send: jest.fn(),
      status: jest.fn().mockReturnThis(),
      locals: {
      }
    };

    userMiddleware(req as any, res as any, nextfn);

    expect(nextfn).not.toBeCalled();
    expect(res.status).toBeCalledWith(403);
    expect(res.send).toBeCalledWith("Missing User Identification");
  });
});
