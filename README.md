<div align="center">
    <img width="200" src="/assets/ofd_logo.svg" alt="OFD Logo">
<h1>OpenFoodDiary</h1>
<h2>Because Food Should Be Simple</h2>
<a href="https://openhealthsuite.com/diary">Project Website</a>
<hr/>
</div>

[![Publish OpenFoodDiary Container](https://github.com/OpenHealthSuite/diary/actions/workflows/publish.yml/badge.svg)](https://github.com/OpenHealthSuite/diary/actions/workflows/publish.yml)
[![Server Build and Test](https://github.com/OpenHealthSuite/diary/actions/workflows/build_and_test_server.yml/badge.svg)](https://github.com/OpenHealthSuite/diary/actions/workflows/build_and_test_server.yml) [![Webapp Build and Test](https://github.com/OpenHealthSuite/diary/actions/workflows/build_and_test_webapp.yml/badge.svg)](https://github.com/OpenHealthSuite/diary/actions/workflows/build_and_test_webapp.yml)

This application is designed as a FOSS webapp for tracking your food. It doesn't want to do anything fancy with the data, nor try and gamify the process - it's intended to be as neutral as possible.

## Just want to try it out?

I host this app for my own use at [https://diary.openhealthsuite.com](https://diary.openhealthsuite.com) - it is behind simple google authentication, and should work fine if you want to try it or use it without setting up your own server.

## Want to host your own?

If you've got a computer running docker at your disposal, you can quickly and easily run a single-user instance of OpenFoodDiary, using sqlite3 as the datastore:

```bash
mkdir openfooddiarydata
docker run -d -v $(pwd)/openfooddiarydata:/app/.sqlite \
  -p 3012:3012 \
  -e PORT="3012" \
  -e OPENFOODDIARY_USERID="my-ofd-userid" \
  --name openfooddiary-instance \
  ghcr.io/openhealthsuite/diary:latest
```

This will start OpenFoodDiary running, on port 3012. You can then access it using your web browser and you're ready to enter your logs - data will be persisted to a sqlite file in the volume mounted directory.

It's entirely possible to run OpenFoodDiary in more complicated configurations - various environment variables for configuration are listed below.

#### Helm?

I keep a trim helm chart in this repository, which with a small amount of manual futzing allows me to deploy updates - can be seen as a starting point to deploying OFD to your own cluster if you have one.

## Running this Repo Locally

This project is a single golang application using SSR with HTMX. Therefore, your only dependency for running the code, is golang installed.

Quickest way to just fire up the server is to run `OPENFOODDIARY_USERID="f1750ac3-d6cc-4981-9466-f1de2ebbad33" go run cmd/server/main.go`

I've included a makefile with the usual suspects in terms of development commands.

There is a docker-compose.yaml file provided for rigging up a quick compose stack with postgres in it - however I'd recommend sticking with the sqlite3 backed storage for dev, and letting the parity tests save you from running postgres for development.

## Environment Variables

### General

- `PORT`: defaults to 8080
  - sets the port OFD will run on
- `OPENFOODDIARY_USERIDHEADER`: defaults to "x-openfooddiary-userid"
  - Denotes the header that will be populated with a user id
- `OPENFOODDIARY_USERID`: no default
  - Denotes userid that will _always_ be populated - intended for dev and single-user modes
- `OPENFOODDIARY_LOGOUT_ENDPOINT`: no default
  - Used in the UI for the logout button, for different auth provider support

### Storage

- `OPENFOODDIARY_POSTGRES_CONNECTION_STRING`: No default
  - If set, the application will attempt to connect and migrate on this string - if it fails, then 
- `OPENFOODDIARY_SQLITE_PATH`: defaults to ".sqlite"
  - Sets the filename/path the sqlite3 database will be stored to
  - note: this location equates to `/app/.sqlite` in the container
