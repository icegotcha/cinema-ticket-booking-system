FROM node:24-alpine AS dependencies

RUN npm install --prefix /opt/seed --omit=dev --ignore-scripts bcryptjs@3.0.3

FROM mongo:9

COPY --from=dependencies /opt/seed/node_modules /opt/seed/node_modules
