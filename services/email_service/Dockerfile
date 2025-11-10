# Dockerfile
FROM python:3.11-alpine AS builder

# Install build dependencies
RUN apk update && apk add --no-cache gcc musl-dev libffi-dev openssl-dev tzdata 

# Set working directory
WORKDIR /app

# Copy requirements file
COPY requirements.txt ./

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy the source code
COPY . .

# Final stage
FROM python:3.11-alpine

# Install runtime dependencies
RUN apk update && apk add --no-cache libffi openssl tzdata

# Create appuser
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy installed packages from builder stage
COPY --from=builder /usr/local/lib/python3.11/site-packages /usr/local/lib/python3.11/site-packages
COPY --from=builder /usr/local/bin /usr/local/bin
COPY --from=builder /app /app

# Expose port
EXPOSE 2525

# Run the application
ENTRYPOINT ["python", "email_service.py"]