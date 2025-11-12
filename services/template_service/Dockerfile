FROM node:20-alpine AS builder

WORKDIR /builder

COPY package*.json ./
RUN npm set progress=false \
    && npm config set fetch-retry-maxtimeout 120000 \
    && npm config set fetch-retry-mintimeout 20000 \
    && npm config set registry https://registry.npmjs.org/ \
    && npm ci --prefer-offline --no-audit

COPY . .
RUN npm run build

FROM node:20-alpine
WORKDIR /app

ENV NODE_ENV=production

COPY --from=builder /builder/node_modules ./node_modules
COPY --from=builder /builder/package*.json ./
COPY --from=builder /builder/dist ./dist

EXPOSE 3000

CMD ["node", "dist/main.js"]
