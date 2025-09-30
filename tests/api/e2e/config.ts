export const TEST_CONFIGURATION = {
  API_HOST: process.env.OPENFOODDIARY_API_HOST ?? "http://localhost:3022",
  USERID_HEADER:
    process.env.OPENFOODDIARY_USERIDHEADER ?? "x-openfooddiary-userid",
};
