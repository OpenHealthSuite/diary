import { userMiddleware } from './UserMiddleware';
import crypto from 'node:crypto';
import { OFDLocals } from './types/Locals';

describe("UserMiddleware", () => {
    test("Happy Path :: By default, takes trusted header and adds to locals", () => {
        const nextfn = jest.fn();
        const testUserId = crypto.randomUUID()
        const req = {
            headers: {
                "X-OpenFoodDiary-UserId": testUserId
            }
        }
        const res = {
            locals: {

            }
        }

        userMiddleware(req as any, res as any, nextfn);

        expect(nextfn).toBeCalledTimes(1);
        expect(nextfn).toBeCalledWith();
        expect((res.locals as OFDLocals).userId).toBe(testUserId);
    })
})