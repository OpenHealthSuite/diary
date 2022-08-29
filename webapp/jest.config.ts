export default {
  transform: {
    '^.+\\.svelte$': ['svelte-jester',
    {
      "preprocess": true
    }],
    '^.+\\.js$': 'babel-jest',
    '^.+\\.ts$': 'ts-jest',
  },
  moduleFileExtensions: ['ts', 'js', 'svelte'],
  testEnvironment: 'jsdom',
  setupFilesAfterEnv: ['<rootDir>/jest-setup.ts']
}