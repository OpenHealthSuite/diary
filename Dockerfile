FROM --platform=$BUILDPLATFORM node:18.7.0 AS server-build
WORKDIR /server

# Get Dependancies
COPY server/package.json server/package-lock.json ./
RUN npm ci

# Copy everything else and build
COPY server ./
RUN npm run build

FROM --platform=$BUILDPLATFORM node:18.7.0 AS server-depns
WORKDIR /server

# Get Dependancies
COPY server/package.json server/package-lock.json ./
RUN npm ci --omit=dev

FROM --platform=$BUILDPLATFORM node:18.7.0 as webapp-build

WORKDIR /webapp

COPY webapp/package.json webapp/package-lock.json ./
RUN npm ci

COPY webapp .
RUN npm run build

# Build runtime image
FROM node:18.7.0-alpine
WORKDIR /app
COPY --from=server-depns /server/node_modules /app/node_modules
COPY --from=server-build /server/dist /app
COPY --from=webapp-build /webapp/dist /app/public
ENTRYPOINT ["node", "index.js"]