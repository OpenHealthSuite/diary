<div align="center">
    <img width="200" src="/assets/ofd_logo.svg" alt="OFD Logo">
<h1>OpenFoodDiary</h1>
<h2>Because Food Should Be Simple</h2>
<a href="https://www.openfooddiary.com/">Project Website</a>
<hr/>
</div>

[![Publish OpenFoodDiary Container](https://github.com/LeeMartin77/openfooddiary/actions/workflows/publish.yml/badge.svg)](https://github.com/LeeMartin77/openfooddiary/actions/workflows/publish.yml)
[![Server Build and Test](https://github.com/LeeMartin77/openfooddiary/actions/workflows/build_and_test_server.yml/badge.svg)](https://github.com/LeeMartin77/openfooddiary/actions/workflows/build_and_test_server.yml) [![Webapp Build and Test](https://github.com/LeeMartin77/openfooddiary/actions/workflows/build_and_test_webapp.yml/badge.svg)](https://github.com/LeeMartin77/openfooddiary/actions/workflows/build_and_test_webapp.yml)

This application is designed as a FOSS webapp for tracking your food. It doesn't want to do anything fancy with the data, nor try and gamify the process - it's intended to be as neutral as possible.

## Just want to try it out?

I host this app for my own use at [https://app.openfooddiary.com](https://app.openfooddiary.com) - it is behind simple google authentication, and should work fine if you want to try it or use it without setting up your own server.

## Want to host your own?

If you've got a computer running podman/docker at your disposal, you can quickly and easily run a single-user instance of OpenFoodDiary, using sqlite3 as the datastore:

```bash
mkdir openfooddiarydata
# podman is interchangeable with docker for this command
podman run -d -v $(pwd)/openfooddiarydata:/app/.sqlite \
  -p 3012:3012 \
  -e OPENFOODDIARY_USERID="my-ofd-userid" \
  --name openfooddiary-instance \
  ghcr.io/leemartin77/openfooddiary:latest
```

This will start OpenFoodDiary running, on port 3012. You can then access it using your web browser and you're ready to enter your logs - data will be persisted to a sqlite file in the volume mounted directory.

It's entirely possible to run OpenFoodDiary in more complicated configurations - various environment variables for configuration are listed below.

#### Helm?

I keep a trim helm chart in this repository, which with a small amount of manual futzing allows me to deploy updates - can be seen as a starting point to deploying OFD to your own cluster if you have one.

## Environment Variables

### General

- `OPENFOODDIARY_PORT`: defaults to 3012
  - sets the port OFD will run on
- `OPENFOODDIARY_USERIDHEADER`: defaults to "x-openfooddiary-userid"
  - Denotes the header that will be populated with a user id
- `OPENFOODDIARY_USERID`: defaults to undefined
  - Denotes userid that will _always_ be populated - intended for dev and single-user modes
- `OPENFOODDIARY_LOGOUT_ENDPOINT`: defaults to `/api/logout`
  - Value that will be returned when the user calls `/api/logout-endpoint`, allowing for different auth providers
- `OPENFOODDIARY_TEMP_DIRECTORY`: defaults to "/tmp"
  - Directory where scratch temp files will be written
- `OPENFOODDIARY_DISABLE_PROMETHEUS`: defaults to undefined
  - disable prometheus metric collection and endpoint

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
