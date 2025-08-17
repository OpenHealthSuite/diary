/** @type {import('ts-jest/dist/types').InitialOptionsTsJest} */
module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'node',
  testMatch: ['**/*.test.ts'],
  testPathIgnorePatterns: ['/node_modules/', '.+\.integration\.test\.ts$'],
  coveragePathIgnorePatterns: [
    "/node_modules/",
    "/src/storage/"
  ]
};