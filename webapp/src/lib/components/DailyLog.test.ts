import { vi } from "vitest";
import type { Mock } from "vitest";
import { render } from "@testing-library/svelte";
import { apiFetch } from "src/lib/utilities";
import type { FoodLogEntry } from "../types/FoodLogEntry";
import crypto from "node:crypto";
import { writable } from "svelte/store";
import type { MetricsConfig } from "src/stores";
import DailyLog from "./DailyLog.svelte";

vi.mock("src/lib/utilities", () => {
  return {
    apiFetch: vi.fn(),
    METRIC_MAX: 999999,
  };
});

vi.mock("src/stores", () => {
  return {
    metricsConfig: writable({
      calories: { label: "Calories", priority: 0 },
    } as MetricsConfig),
    logUpdated: writable(new Date().toISOString()),
  };
});

describe("Daily Log", () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it("it requests based on input day :: midnight-midnight", async () => {
    const testDate = new Date(2017, 8, 10, 12, 0, 0);
    const data = [];

    const response = {
      status: 200,
      json: vi.fn().mockResolvedValue(data),
    };

    (apiFetch as Mock<any[], any>).mockResolvedValue(response);

    render(DailyLog, {
      day: testDate,
    });

    const expectedStart = new Date(testDate.toISOString().split("T")[0]);
    const expectedEnd = new Date(testDate.toISOString().split("T")[0]);
    expectedEnd.setDate(expectedEnd.getDate() + 1);

    expect(apiFetch).toBeCalledWith(
      `/logs?startDate=${expectedStart.toISOString()}&endDate=${expectedEnd.toISOString()}`
    );
  });

  describe("Loading and Error logic", () => {
    afterEach(() => {
      vi.clearAllMocks();
    });

    it("it has loading indicator while loading, gone when resolved", async () => {
      const testDate = new Date(2017, 8, 10, 12, 0, 0);

      const data = [];

      const response = {
        status: 200,
        json: vi.fn().mockResolvedValue(data),
      };

      let promiseResolve;

      (apiFetch as Mock<any[], any>).mockReturnValue(
        new Promise((resolve) => (promiseResolve = resolve))
      );
      const { getByTestId, queryByTestId } = render(DailyLog, {
        day: testDate,
      });

      expect(getByTestId("loading-indicator")).toBeInTheDocument();

      promiseResolve(response);

      await new Promise(process.nextTick);

      expect(queryByTestId("loading-indicator")).not.toBeInTheDocument();
    });

    it("it has loading indicator while loading, gone when rejected", async () => {
      const testDate = new Date(2017, 8, 10, 12, 0, 0);

      const response = {};

      let promiseReject;

      (apiFetch as Mock<any[], any>).mockReturnValue(
        new Promise((resolve, reject) => (promiseReject = reject))
      );
      const { getByTestId, queryByTestId } = render(DailyLog, {
        day: testDate,
      });

      expect(getByTestId("loading-indicator")).toBeInTheDocument();

      promiseReject(response);

      await new Promise(process.nextTick);

      expect(queryByTestId("loading-indicator")).not.toBeInTheDocument();
    });

    const errorCodes = [400, 403, 500];

    it.each(errorCodes)("shows error message on non-200", async (status) => {
      const testDate = new Date(2017, 8, 10, 12, 0, 0);

      (apiFetch as Mock<any[], any>).mockResolvedValue({
        status,
        json: vi.fn().mockResolvedValue([]),
      });
      const { getByTestId } = render(DailyLog, {
        day: testDate,
      });

      await new Promise(process.nextTick);

      expect(getByTestId("error-indicator")).toBeInTheDocument();
    });

    it("shows error message on reject", async () => {
      const testDate = new Date(2017, 8, 10, 12, 0, 0);

      (apiFetch as Mock<any[], any>).mockRejectedValue({});
      const { getByTestId } = render(DailyLog, {
        day: testDate,
      });

      await new Promise(process.nextTick);

      expect(getByTestId("error-indicator")).toBeInTheDocument();
    });
  });

  describe("Data Display", () => {
    afterEach(() => {
      vi.clearAllMocks();
    });

    it("Tells the user when no logs", async () => {
      const testDate = new Date(2017, 8, 10, 12, 0, 0);

      const data = [];

      const response = {
        status: 200,
        json: vi.fn().mockResolvedValue(data),
      };

      (apiFetch as Mock<any[], any>).mockResolvedValue(response);

      const { getByText } = render(DailyLog, {
        day: testDate,
      });

      await new Promise(process.nextTick);

      expect(getByText("No Logs Entered for this day")).toBeInTheDocument();
    });

    it("Shows the user the logs, in start time order, with calories", async () => {
      const testDate = new Date(2017, 8, 10, 12, 0, 0);

      const data: FoodLogEntry[] = [
        {
          id: crypto.randomUUID(),
          name: "Food Log Two",
          labels: ["label one", "label two"],
          time: {
            start: new Date(2017, 8, 10, 11, 0, 0),
            end: new Date(2017, 8, 10, 11, 30, 0),
          },
          metrics: {
            calories: 500,
          },
        },
        {
          id: crypto.randomUUID(),
          name: "Food Log One",
          labels: ["label one", "label two"],
          time: {
            start: new Date(2017, 8, 10, 10, 20, 0),
            end: new Date(2017, 8, 10, 10, 50, 0),
          },
          metrics: {
            calories: 500,
          },
        },
        {
          id: crypto.randomUUID(),
          name: "Food Log Three",
          labels: ["label one", "label two"],
          time: {
            start: new Date(2017, 8, 10, 11, 15, 0),
            end: new Date(2017, 8, 10, 11, 30, 0),
          },
          metrics: {
            calories: 500,
          },
        },
      ];

      const response = {
        status: 200,
        json: vi.fn().mockResolvedValue(data),
      };

      (apiFetch as Mock<any[], any>).mockResolvedValue(response);

      const { getByTestId } = render(DailyLog, {
        day: testDate,
      });

      await new Promise(process.nextTick);

      data
        .sort((a, b) => a.time.start.getTime() - b.time.start.getTime())
        .map((x) => [x.name, x.metrics.calories])
        .forEach(([name, calcount], i) => {
          const element = getByTestId(`foodlog-${i}`);
          expect(element).toBeInTheDocument();
          expect(element).toHaveTextContent(name as string);
          const calories = getByTestId(`foodlog-${i}-calories`);
          expect(calories).toBeInTheDocument();
          expect(calories).toHaveTextContent(calcount.toString());
        });
    });
  });
});
