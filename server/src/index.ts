import express from "express";
import qs from "qs";
import { FoodStorageRouter, ConfigurationRouter } from "./handlers";
import { userMiddleware } from "./middlewares";
import { STORAGE } from "./storage";
import { prometheusSetup } from "./middlewares/PrometheusMiddleware";

const port = process.env.OPENFOODDIARY_PORT || 3012;

const app = express();

app.settings["query parser"] = qs.parse;

app.use(express.json());
app.use(express.static("./public"));

if (!process.env.OPENFOODDIARY_DISABLE_PROMETHEUS) {
  prometheusSetup(app);
}

app.get("/api/health", (req, res) => {
  res.send();
});

app.get("/api/logout-endpoint", (req, res) => {
  res.send({
    url: process.env.OPENFOODDIARY_LOGOUT_ENDPOINT ?? "/api/logout",
  });
});

app.use(userMiddleware);

app.use("/api", FoodStorageRouter);
app.use("/api", ConfigurationRouter);

app.get("/*", (req, res) => {
  res.sendFile("index.html", { root: "./public" });
});

app.get("/api/*", (req, res) => {
  res.sendStatus(404);
});

STORAGE.setupDatabase().then(() => {
  const running = app.listen(port, () => {
    console.log(`OpenFoodDiary listening on port ${port}`);

    const cleanup = () => {
      STORAGE.shutdownDatabase();
      running.close();
    };

    process.on("SIGTERM", cleanup);
    process.on("SIGINT", cleanup);
    process.on("exit", cleanup);
  });
});
