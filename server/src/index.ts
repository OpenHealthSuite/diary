import express from "express";
import qs from "qs";
import { FoodStorageRouter } from "./handlers";
import { userMiddleware } from "./middlewares";
import { STORAGE } from "./storage";

const port = process.env.OPENFOODDIARY_PORT || 3012;

const app = express();

app.settings["query parser"] = qs.parse;

app.use(express.json());
app.use(express.static("./public"));

app.get("/api/health", (req, res) => {
  res.send();
});

app.use(userMiddleware);

app.use("/api", FoodStorageRouter);

STORAGE.setupDatabase().then(() => {
  const running = app.listen(port, () => {
    console.log(`Example app listening on port ${port}`);

    const cleanup = () => {
      STORAGE.shutdownDatabase();
      running.close();
    };

    process.on("SIGTERM", cleanup);
    process.on("SIGINT", cleanup);
    process.on("exit", cleanup);
  });
});
