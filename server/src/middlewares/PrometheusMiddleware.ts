import { NextFunction, Response, Request } from "express";
import { PROM_PREFIX } from "../config";
import promclient from "prom-client";
import { Express } from "express";

promclient.collectDefaultMetrics({
  prefix: PROM_PREFIX,
});

const requests = new promclient.Counter({
  name: `${PROM_PREFIX}_http_requests`,
  help: "HTTP Request count with statuses",
  labelNames: ["method", "statusCode"],
});

export function prometheusSetup(app: Express) {
  app.use(promMiddleware);
  app.get("/api/metrics", async (req, res) => {
    res.setHeader("Content-Type", promclient.register.contentType);
    res.send(await promclient.register.metrics());
  });
  return app;
}

export function promMiddleware(
  req: Request,
  res: Response,
  next: NextFunction
) {
  requests.labels(req.method, res.statusCode.toString()).inc();
  next();
}
