import express, { Request, Response, Router } from "express";
import { StorageError, STORAGE } from "../storage";
import { OFDLocals } from "../middlewares";
import { Configuration } from "../types";
const configurationStorage = STORAGE.configuration;

export function buildRouter (router: Router): Router {
  return router
    .get("/config/:configId", getConfigHandler)
    .post("/config/:configId", saveConfigHandler);
}

export const ConfigurationRouter = buildRouter(express.Router());

const storageErrorHandler = (err: StorageError, res: Response) => {
  switch (err.errorType) {
    case "notfound":
      res.sendStatus(404);
      break;
    case "validation":
      res.sendStatus(400);
      break;
    case "system":
      res.sendStatus(500);
      break;
  }
};

function getConfigHandler (req: Request, res: Response & { locals: OFDLocals }) {
  const { configId } = req.params;
  configurationStorage
    .retrieveUserConfiguration(res.locals.userId, configId as any)
    .then((configRes) => {
      configRes
        // eslint-disable-next-line array-callback-return
        .map((config) => {
          res.contentType("json").send(config);
        })
        .mapErr((err) => storageErrorHandler(err, res));
    });
}

function saveConfigHandler (
  req: Request,
  res: Response & { locals: OFDLocals }
) {
  const { configId } = req.params;
  const config: Configuration = {
    id: configId as any,
    value: req.body
  };
  configurationStorage
    .storeConfiguration(res.locals.userId, config)
    .then((configRes) => {
      configRes
        // eslint-disable-next-line array-callback-return
        .map(() => {
          res.sendStatus(200);
        })
        .mapErr((err) => storageErrorHandler(err, res));
    });
}
