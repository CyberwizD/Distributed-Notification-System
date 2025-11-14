<<<<<<< HEAD
FROM javascript:20-alpine AS builder

# Install git and ca-certificates
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# Create appuser
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /builder

# Copy package files
COPY package.json package-lock.json ./

# Install dependencies
RUN npm install

# Copy the source code
COPY . .

# Build the application
RUN npx prisma generate
RUN npm run build

# Final stage
FROM alpine:latest

# Install ca-certificates
RUN apk update && apk add --no-cache ca-certificates tzdata && update-ca-certificates

# Create appuser
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy the built files from builder stage
COPY --from=builder /builder/dist /app
COPY --from=builder /builder/node_modules /app/node_modules
COPY --from=builder /builder/package.json /app/package.json
COPY --from=builder /builder/package-lock.json /app/package-lock.json

# EXPOSE port
EXPOSE 3000

# Run the application
ENTRYPOINT ["node", "index.js"]
=======
FROM node:20-alpine AS builder

WORKDIR /builder

COPY package*.json ./
RUN npm set progress=false \
    && npm config set fund false \
    && npm config set update-notifier false \
    && npm config set fetch-retry-maxtimeout 120000 \
    && npm config set fetch-retry-mintimeout 20000 \
    && (npm ci --prefer-offline --no-audit --no-fund || npm ci --no-audit --no-fund)

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
>>>>>>> 498ba6a6833b684248f4c02c6506ccb60404fd84
