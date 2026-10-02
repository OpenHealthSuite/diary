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

#### Kubernetes?

I keep plain kustomize manifests in `kustomize/` in this repository, which allow me to deploy updates with a small amount of manual futzing - can be seen as a starting point to deploying OFD to your own cluster if you have one.

The base is configured for the Oauth2 login flow, and expects a `openfooddiary-session` secret in the target namespace containing a `session-secret` key:

```bash
kubectl create secret generic openfooddiary-session --from-literal=session-secret="$(openssl rand -base64 32)" -n diary
kubectl apply -k kustomize/base
```

If you'd rather run in single-user mode (no authentication, no redis, no session secret), there's an overlay for it - set `OPENFOODDIARY_USERID` in `kustomize/overlays/single-user/deployment-patch.yml` and then:

```bash
kubectl apply -k kustomize/overlays/single-user
```

## Running this Repo Locally

This project is a single golang application using SSR with HTMX. Therefore, your only dependency for running the code, is golang installed.

Quickest way to just fire up the server is to run `OPENFOODDIARY_USERID="f1750ac3-d6cc-4981-9466-f1de2ebbad33" go run cmd/server/main.go`

I've included a makefile with the usual suspects in terms of development commands.

There is a docker-compose.yaml file provided for rigging up a quick compose stack with postgres in it - however I'd recommend sticking with the sqlite3 backed storage for dev, and letting the parity tests save you from running postgres for development.

## Environment Variables

### General

- `PORT`: defaults to 8080
  - sets the port OFD will run on
- `OPENFOODDIARY_USERID`: no default
  - Denotes userid that will _always_ be populated - intended for dev and single-user modes
  - If set, authentication is bypassed entirely and the Oauth2 variables below are ignored
- `OPENFOODDIARY_LOGOUT_ENDPOINT`: no default
  - Overrides where the UI's logout button points. Defaults to the app's own logout handler (`/auth/logout`) when the Oauth2 flow is in use

### Authentication

If `OPENFOODDIARY_USERID` is not set, the app expects to be configured against an OIDC provider.

- `OPENFOODDIARY_OAUTH2_ISSUER`: no default, required
  - The provider's issuer url, e.g. `https://dex.example.com`
- `OPENFOODDIARY_OAUTH2_CLIENT_ID`: no default, required
  - The client id registered with the provider
- `OPENFOODDIARY_OAUTH2_CLIENT_SECRET`: no default
  - The client secret
- `OPENFOODDIARY_OAUTH2_REDIRECT_URL`: no default
  - The callback url to send to the provider. If unset, it is derived from the incoming request, honouring `X-Forwarded-Proto` and `X-Forwarded-Host`
  - The default resolves to `https://<your-host>/auth/callback` - register that with your provider
- `OPENFOODDIARY_OAUTH2_SKIP_ISSUER_VERIFICATION`: defaults to false
  - Disables the id_token issuer/audience checks. Only useful against a dev provider

The app only requests the `openid` scope, and keys your data off the `sub` claim from the returned id_token - whatever stable identifier your provider puts there, typically a UUID with dex.

### Sessions

Sessions are stored in redis, which the app requires when running the Oauth2 flow.

- `OPENFOODDIARY_SESSION_REDIS_URL`: no default, required for the Oauth2 flow
  - e.g. `redis://user:password@redis:6379/0`
- `OPENFOODDIARY_SESSION_SECRET`: no default, required for the Oauth2 flow
  - Authenticates and encrypts the session cookie.
- `OPENFOODDIARY_SESSION_COOKIE_NAME`: defaults to "openfooddiary-session"
- `OPENFOODDIARY_SESSION_MAX_AGE`: defaults to 86400
  - Session lifetime in seconds
- `OPENFOODDIARY_SESSION_SECURE`: defaults to true

### Storage

- `OPENFOODDIARY_POSTGRES_CONNECTION_STRING`: No default
  - If set, the application will attempt to connect and migrate on this string - if it fails, then 
- `OPENFOODDIARY_SQLITE_PATH`: defaults to ".sqlite"
  - Sets the filename/path the sqlite3 database will be stored to
  - note: this location equates to `/app/.sqlite` in the container

## Running the tests

```bash
make test
```

The storage tests and the OIDC login flow tests both spin up throwaway postgres and redis containers via testcontainers, so docker needs to be running.
