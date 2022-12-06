# OpenFoodDiary

[![Publish OpenFoodDiary Container](https://github.com/LeeMartin77/openfooddiary/actions/workflows/publish.yml/badge.svg)](https://github.com/LeeMartin77/openfooddiary/actions/workflows/publish.yml)
[![Server Build and Test](https://github.com/LeeMartin77/openfooddiary/actions/workflows/build_and_test_server.yml/badge.svg)](https://github.com/LeeMartin77/openfooddiary/actions/workflows/build_and_test_server.yml) [![Webapp Build and Test](https://github.com/LeeMartin77/openfooddiary/actions/workflows/build_and_test_webapp.yml/badge.svg)](https://github.com/LeeMartin77/openfooddiary/actions/workflows/build_and_test_webapp.yml)

This application is designed as a FOSS webapp for tracking your food. It doesn't want to do anything fancy with the data, nor try and gamify the process - it's intended to be as neutral as possible.

## Environment Variables

### General

- `OPENFOODDIARY_PORT`: defaults to 3012
  - sets the port OFD will run on
- `OPENFOODDIARY_USERIDHEADER`: defaults to "x-openfooddiary-userid"
  - Denotes the header that will be populated with a user id
- `OPENFOODDIARY_USERID`: defaults to undefined
  - Denotes userid that will _always_ be populated - intended for dev and single-user modes

### Storage

- `OPENFOODDIARY_STORAGE_PROVIDER`: defaults to undefined
  - Sets the storage provider, OFD falls back to sqlite3 if none is defined
  - options: `cassandra`
- `OPENFOODDIARY_SQLITE3_FILENAME`: defaults to ".sqlite/openfooddiary.sqlite"
  - Sets the filename/path the sqlite3 database will be stored to
  - note: this location equates to `/app/.sqlite/openfooddiary.sqlite` in the container
- `OPENFOODDIARY_CASSANDRA_`...
  - ...`CONTACT_POINTS`: defaults to "localhost:9042"
  - ...`LOCALDATACENTER`: defaults to "datacenter1"
  - ...`USER`: defaults to "cassandra"
  - ...`PASSWORD`: defaults to "cassandra"
  - note: Defaults designed to work with setup script in `.development`
