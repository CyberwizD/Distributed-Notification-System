FROM node:20-alpine AS builder

WORKDIR /builder

COPY package*.json ./
RUN npm ci

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
