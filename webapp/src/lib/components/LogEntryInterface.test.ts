import LogEntryInterface from './LogEntryInterface.svelte'
import { vi } from 'vitest';
import type { Mock } from 'vitest';
import { render, fireEvent } from '@testing-library/svelte'
import { apiFetch } from 'src/lib/utilities'
import type { FoodLogEntry } from '../types/FoodLogEntry';
import crypto from 'node:crypto';

vi.mock('src/lib/utilities', () => {
    return { apiFetch: vi.fn() }
})

describe("Create Log", () => {
    afterEach(() => {
      vi.clearAllMocks()
    })
    
    it('start :: it renders form with empty fields and current time', async () => {
      const date = new Date(2018, 1, 1, 13, 15, 23)
      vi.setSystemTime(date)

      const { getByLabelText } = render(LogEntryInterface)

      const name = getByLabelText("Log Name", { exact: false });
      // Really just ended up fighting the custom date/time pickers here
      // const dateInput = getByTestId("date-picker");
      // const time = getByLabelText("Time");
      const duration = getByLabelText("Duration (minutes)", { exact: false });
      const calories = getByLabelText("Calories", { exact: false });

      expect(name).toBeInTheDocument();
      // expect(dateInput).toBeInTheDocument();
      // expect(time).toBeInTheDocument();
      expect(duration).toBeInTheDocument();
      expect(calories).toBeInTheDocument();

      expect(name).toHaveValue("")
      expect(duration).toHaveValue(1);
      expect(calories).toHaveValue(0);
    })

    it('validation :: no name given, submit is not enabled', async () => {
      const date = new Date(2018, 1, 1, 13, 15, 23)
      vi.setSystemTime(date)

      const { getByLabelText, getByText } = render(LogEntryInterface)

      const name = getByLabelText("Log Name", { exact: false });
      expect(name).toBeInTheDocument()
      expect(name).toHaveValue("")

      const submitButton = getByText("Submit")
      expect(submitButton).toBeInTheDocument();
      expect(submitButton).toBeDisabled();
      await fireEvent.click(submitButton)
      expect(apiFetch).not.toBeCalled();
    })

    it('Happy Path :: set name, calories, duration, sends expected request, outputs success', async () => {
      const date = new Date(2018, 1, 1, 13, 15, 23)
      vi.setSystemTime(date)

      const logNameInput = "My Test Log";
      const durationInput = 10;
      const caloriesInput = 500;
      const endDate = new Date(date)
      endDate.setMinutes(endDate.getMinutes() + durationInput)

      const expectedRequest = {
          name: logNameInput,
          labels: new Set(),
          time: {
              start: date.toISOString(),
              end: endDate.toISOString()
          },
          metrics: {
              calories: caloriesInput
          }
      };

      const data = crypto.randomUUID()

      const response = {
          status: 200,
          json: vi.fn().mockResolvedValue(data)
      };

      (apiFetch as Mock<any[], any>).mockResolvedValue(response)

      const { getByLabelText, getByText, component } = render(LogEntryInterface)

      let componentOutput = undefined;

      component.$on('success', e => { componentOutput = e.detail })

      const name = getByLabelText("Log Name", { exact: false });
      const duration = getByLabelText("Duration (minutes)", { exact: false });
      const calories = getByLabelText("Calories", { exact: false });

      fireEvent.input(name, {target: {value: logNameInput}})
      fireEvent.input(duration, {target: {value: durationInput}})
      fireEvent.input(calories, {target: {value: caloriesInput}})

      await new Promise(process.nextTick);

      const submitButton = getByText("Submit")
      expect(submitButton).toBeInTheDocument();
      expect(submitButton).not.toBeDisabled();
      await fireEvent.click(submitButton)
      expect(apiFetch).toBeCalledWith("/logs", {
        method: 'POST',
        body: JSON.stringify(expectedRequest)
      });

      await new Promise(process.nextTick);
      
      expect(componentOutput).toBe(data)
    })

    const errorCodes = [400, 403, 500]

    it.each(errorCodes)('Happy Path :: set name, calories, duration, sends expected request, non 200 response, outputs error', async (errorCode) => {
      const date = new Date(2018, 1, 1, 13, 15, 23)
      vi.setSystemTime(date)

      const logNameInput = "My Test Log";
      const durationInput = 10;
      const caloriesInput = 500;
      const endDate = new Date(date)
      endDate.setMinutes(endDate.getMinutes() + durationInput)

      const expectedRequest = {
          name: logNameInput,
          labels: new Set(),
          time: {
              start: date.toISOString(),
              end: endDate.toISOString()
          },
          metrics: {
              calories: caloriesInput
          }
      };

      const data = crypto.randomUUID()

      const response = {
          status: errorCode,
          text: vi.fn().mockResolvedValue(data)
      };

      (apiFetch as Mock<any[], any>).mockResolvedValue(response)

      const { getByLabelText, getByText, component } = render(LogEntryInterface)

      let componentOutput = undefined;

      component.$on('error', e => { componentOutput = e.detail })

      const name = getByLabelText("Log Name", { exact: false });
      const duration = getByLabelText("Duration (minutes)", { exact: false });
      const calories = getByLabelText("Calories", { exact: false });

      fireEvent.input(name, {target: {value: logNameInput}})
      fireEvent.input(duration, {target: {value: durationInput}})
      fireEvent.input(calories, {target: {value: caloriesInput}})

      await new Promise(process.nextTick);

      const submitButton = getByText("Submit")
      expect(submitButton).toBeInTheDocument();
      expect(submitButton).not.toBeDisabled();
      await fireEvent.click(submitButton)
      expect(apiFetch).toBeCalledWith("/logs", {
        method: 'POST',
        body: JSON.stringify(expectedRequest)
      });

      await new Promise(process.nextTick);
      
      expect(componentOutput).toBe(data)
    })

    it('Unhappy Path :: set name, calories, duration, sends expected request, outputs error', async () => {
      const date = new Date(2018, 1, 1, 13, 15, 23)
      vi.setSystemTime(date)

      const logNameInput = "My Test Log";
      const durationInput = 10;
      const caloriesInput = 500;
      const endDate = new Date(date)
      endDate.setMinutes(endDate.getMinutes() + durationInput)

      const expectedRequest = {
          name: logNameInput,
          labels: new Set(),
          time: {
              start: date.toISOString(),
              end: endDate.toISOString()
          },
          metrics: {
              calories: caloriesInput
          }
      };

      const data = "Something went wrong";

      (apiFetch as Mock<any[], any>).mockRejectedValue(data)

      const { getByLabelText, getByText, component } = render(LogEntryInterface)

      let componentOutput = undefined;

      component.$on('error', e => { componentOutput = e.detail })

      const name = getByLabelText("Log Name", { exact: false });
      const duration = getByLabelText("Duration (minutes)", { exact: false });
      const calories = getByLabelText("Calories", { exact: false });

      fireEvent.input(name, {target: {value: logNameInput}})
      fireEvent.input(duration, {target: {value: durationInput}})
      fireEvent.input(calories, {target: {value: caloriesInput}})

      await new Promise(process.nextTick);

      const submitButton = getByText("Submit")
      expect(submitButton).toBeInTheDocument();
      expect(submitButton).not.toBeDisabled();
      await fireEvent.click(submitButton)
      expect(apiFetch).toBeCalledWith("/logs", {
        method: 'POST',
        body: JSON.stringify(expectedRequest)
      });

      await new Promise(process.nextTick);
      
      expect(componentOutput).toBe("An unknown error occured")
    })
  })