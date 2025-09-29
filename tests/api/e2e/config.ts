export const TEST_CONFIGURATION = {
  API_HOST: process.env.OPENFOODDIARY_API_HOST ?? "http://localhost:8080",
  USERID_HEADER:
    process.env.OPENFOODDIARY_USERIDHEADER ?? "x-openfooddiary-userid",
};
