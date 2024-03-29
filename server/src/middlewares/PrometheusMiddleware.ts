import { NextFunction, Response, Request, Express } from "express";
import { PROM_PREFIX } from "../config";
import promclient from "prom-client";

promclient.collectDefaultMetrics({
  prefix: PROM_PREFIX
});

const requests = new promclient.Counter({
  name: `${PROM_PREFIX}http_requests`,
  help: "HTTP Request count with statuses",
  labelNames: ["type", "method", "statusCode"]
});

const METRICS_ENDPOINT = "/metrics";

export function prometheusSetup (app: Express) {
  app.use(promMiddleware);
  app.get(METRICS_ENDPOINT, async (req, res) => {
    res.setHeader("Content-Type", promclient.register.contentType);
    res.send(await promclient.register.metrics());
  });
  return app;
}

export function promMiddleware (
  req: Request,
  res: Response,
  next: NextFunction,
  reqcounter = requests
) {
  if (req.path !== METRICS_ENDPOINT) {
    reqcounter
      .labels(
        req.path.startsWith("/api") ? "api" : "client",
        req.method,
        res.statusCode.toString()
      )
      .inc();
  }
  next();
}
