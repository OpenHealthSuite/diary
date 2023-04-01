import { ConfigurationRouter } from "./ConfigHandlers";
import express, { Express } from "express";
import { STORAGE } from "../storage";
import {
  NotFoundError,
  StorageError,
  SystemError,
  ValidationError
} from "../storage/types/StorageErrors";
import { err, ok } from "neverthrow";
import request from "supertest";
import qs from "qs";

jest.mock("../storage", () => {
  return {
    STORAGE: {
      configuration: {
        retrieveUserConfiguration: jest.fn(),
        storeConfiguration: jest.fn()
      }
    }
  };
});

const errors: { error: StorageError; statusCode: number }[] = [
  { error: new ValidationError(), statusCode: 400 },
  { error: new NotFoundError(), statusCode: 404 },
  { error: new SystemError(), statusCode: 500 }
];

const FAKE_ROOT = "/testapi";
const FAKE_USER_ID = "some-user-id";
const fakeUserMiddleware = (req: any, res: any, next: any) => {
  res.locals.userId = FAKE_USER_ID;
  next();
};
describe("ConfigHandlers", () => {
  let app: Express;
  beforeEach(() => {
    app = express();
    app.settings["query parser"] = qs.parse;
    app.use(express.json());
    app.use(fakeUserMiddleware);
    app.use("/testapi", ConfigurationRouter);
  });

  describe("GET config", () => {
    beforeEach(() => {
      jest.resetAllMocks();
    });
    test("HappyPath :: Storage returns config, returns", async () => {
      const fakeConfig = { whoami: "fakeconfig" };

      (
        STORAGE.configuration.retrieveUserConfiguration as jest.Mock
      ).mockResolvedValue(ok(fakeConfig));

      const res = await request(app)
        .get(FAKE_ROOT + "/config/metrics")
        .expect("Content-Type", /json/)
        .expect(200);

      expect(res.body).toEqual(fakeConfig);
      expect(
        STORAGE.configuration.retrieveUserConfiguration as jest.Mock
      ).toBeCalledWith(FAKE_USER_ID, "metrics");
    });
    test.each(errors)(
      "Storage Error $error.errorType :: Returns $statusCode code",
      async ({ error, statusCode }) => {
        (
          STORAGE.configuration.retrieveUserConfiguration as jest.Mock
        ).mockResolvedValue(err(error));

        await request(app)
          .get(FAKE_ROOT + "/config/metrics")
          .expect(statusCode);

        expect(
          STORAGE.configuration.retrieveUserConfiguration as jest.Mock
        ).toBeCalledWith(FAKE_USER_ID, "metrics");
      }
    );
  });

  describe("POST config", () => {
    beforeEach(() => {
      jest.resetAllMocks();
    });
    test("HappyPath :: Storage takes in new config", async () => {
      const fakeConfig = { whoami: "fakeconfig" };

      (STORAGE.configuration.storeConfiguration as jest.Mock).mockResolvedValue(
        ok("metrics")
      );

      await request(app)
        .post(FAKE_ROOT + "/config/metrics")
        .send(fakeConfig)
        .set("Accept", "application/json")
        .expect(200);

      expect(
        STORAGE.configuration.storeConfiguration as jest.Mock
      ).toBeCalledWith(FAKE_USER_ID, { id: "metrics", value: fakeConfig });
    });
    test.each(errors)(
      "Storage Error $error.errorType :: Returns $statusCode code",
      async ({ error, statusCode }) => {
        const fakeConfig = { whoami: "fakeconfig" };

        (
          STORAGE.configuration.storeConfiguration as jest.Mock
        ).mockResolvedValue(err(error));
        await request(app)
          .post(FAKE_ROOT + "/config/metrics")
          .send(fakeConfig)
          .expect(statusCode);

        expect(
          STORAGE.configuration.storeConfiguration as jest.Mock
        ).toBeCalledWith(FAKE_USER_ID, { id: "metrics", value: fakeConfig });
      }
    );
  });
});
