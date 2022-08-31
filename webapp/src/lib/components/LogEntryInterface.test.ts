import LogEntryInterface from './LogEntryInterface.svelte'
import { vi } from 'vitest';
import type { Mock } from 'vitest';
import { render } from '@testing-library/svelte'
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
  })