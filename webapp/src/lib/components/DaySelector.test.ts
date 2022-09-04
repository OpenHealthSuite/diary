import DaySelector from './DaySelector.svelte'
import { vi } from 'vitest';
import { render, fireEvent } from '@testing-library/svelte'

describe("DaySelector", () => {
    it("without date prop :: uses system time as base", async () => {
        const date = new Date(2018, 1, 1, 13, 15, 23)
        vi.setSystemTime(date)

        const { getByLabelText, getByText, getByTestId, component } = render(DaySelector)

        let componentOutput = undefined;
  
        component.$on('dateChange', e => { componentOutput = e.detail })

        const forwardButton = getByTestId('forward-button')
        const backwardButton = getByTestId('backward-button')

        const movingDate = new Date(date);
        movingDate.setDate(movingDate.getDate() + 1)

        await fireEvent.click(forwardButton)

        // await new Promise(process.nextTick);

        expect((componentOutput as Date).toISOString().split('T')[0])
            .toBe(movingDate.toISOString().split('T')[0]);

        movingDate.setDate(movingDate.getDate() - 2)

        await fireEvent.click(backwardButton)
        await fireEvent.click(backwardButton)

        // await new Promise(process.nextTick);

        expect((componentOutput as Date).toISOString().split('T')[0])
            .toBe(movingDate.toISOString().split('T')[0]);
  
    })
})