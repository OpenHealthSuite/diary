import { NextFunction, Response, Request } from "express";

export function userMiddleware(req: Request, res: Response, next: NextFunction) {
    res.locals.userId = req.headers["X-OpenFoodDiary-UserId"];
    next();
}