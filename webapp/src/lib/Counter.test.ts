import '@testing-library/jest-dom'
import Counter from './Counter.svelte'
import { render, fireEvent } from '@testing-library/svelte'

it('it works', async () => {
  const { getByTestId } = render(Counter, {
    count: 10,
  })

  // Not ideal that this didn't work out the gate
  const counter = getByTestId('counter-button')

  expect(counter).toHaveTextContent('10')

  await fireEvent.click(counter)
  await fireEvent.click(counter)

  expect(counter).toHaveTextContent('12')
  await fireEvent.click(counter)
  await fireEvent.click(counter)

  expect(counter).toHaveTextContent('14')
})