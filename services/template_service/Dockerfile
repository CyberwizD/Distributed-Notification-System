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
