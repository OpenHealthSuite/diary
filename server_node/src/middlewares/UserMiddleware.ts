import { NextFunction, Response, Request } from "express";

export function userMiddleware (req: Request, res: Response, next: NextFunction) {
  const USER_ID_HEADER = process.env.OPENFOODDIARY_USERIDHEADER ?? "x-openfooddiary-userid";
  if (process.env.OPENFOODDIARY_USERID || req.headers[USER_ID_HEADER]) {
    res.locals.userId = process.env.OPENFOODDIARY_USERID ?? req.headers[USER_ID_HEADER];
    next();
    return;
  }
  res.status(403).send("Missing User Identification");
}
