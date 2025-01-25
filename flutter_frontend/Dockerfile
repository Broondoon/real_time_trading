# Dockerfile initially generated via Gemini 2.0 Flash Experimental

# Stage 1: Build the Flutter Web Application
FROM ubuntu:latest AS builder

# Install necessary tools
RUN apt-get update && apt-get install -y curl git wget unzip xz-utils

# Install Flutter SDK
RUN wget https://storage.googleapis.com/flutter_infra_release/releases/stable/linux/flutter_linux_3.27.3-stable.tar.xz
RUN tar xf flutter_linux_3.27.3-stable.tar.xz

# Create a non-root user and group
RUN groupadd -r flutter && useradd -r -g flutter -d /home/flutter -s /bin/bash flutter

RUN mkdir -p /home/flutter
RUN chown -R flutter:flutter /flutter

# Ensure the home directory is owned by the flutter user
RUN chown -R flutter:flutter /home/flutter

ENV PATH="/flutter/bin:${PATH}"

# Switch to the non-root user
USER flutter

# Create work directory and copy project
WORKDIR /app
COPY --chown=flutter:flutter . .

# Download dependencies and build
RUN flutter pub get
RUN flutter build web --release --web-renderer html

# Stage 2: Create the Production Image with Nginx
FROM nginx:alpine

# Copy the built web app from the builder stage
COPY --from=builder /app/build/web /usr/share/nginx/html

# Expose port 80
EXPOSE 80

# Start Nginx
CMD ["nginx", "-g", "daemon off;"]