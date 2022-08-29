import DailyLog from './DailyLog.svelte'
import { vi } from 'vitest';
import { render, fireEvent } from '@testing-library/svelte'
import { apiFetch } from 'src/lib/utilities'

vi.mock('src/lib/utilities', () => {
    return { apiFetch: vi.fn() }
})

describe("Daily Log", () => {
    afterEach(() => {
      vi.clearAllMocks()
    })
    
    it('it requests based on input day :: midnight-midnight', async () => {
        const testDate = new Date(2017, 8, 10, 12, 0, 0);


        const { getByTestId } = render(DailyLog, {
            day: testDate
        })

        const expectedStart = (new Date(testDate.toISOString().split("T")[0]))
        const expectedEnd = (new Date(testDate.toISOString().split("T")[0]))
        expectedEnd.setDate(expectedEnd.getDate() + 1)

        expect(apiFetch).toBeCalledWith(`/logs?startDate=${expectedStart.toISOString()}&endDate=${expectedEnd.toISOString()}`)
    })    
})
