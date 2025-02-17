# Dockerfile initially generated via Gemini 2.0 Flash Experimental

# Stage 1: Build the Flutter Web Application
#FROM ubuntu:latest AS builder
FROM debian:latest AS build-env

# Install necessary tools
RUN apt-get update && apt-get install -y curl git unzip
# removed: wget xz-utils

ARG SDK=/usr/local/flutter
ARG VER=3.27.3
ARG APP=/app/

# Install Flutter SDK
#RUN wget https://storage.googleapis.com/flutter_infra_release/releases/stable/linux/flutter_linux_3.27.3-stable.tar.xz
#RUN tar xf flutter_linux_3.27.3-stable.tar.xz
RUN git clone https://github.com/flutter/flutter.git $SDK
RUN cd $SDK && git fetch && git checkout $SDK

# Create a non-root user and group
#RUN groupadd -r flutter && useradd -r -g flutter -d /home/flutter -s /bin/bash flutter

#RUN mkdir -p /home/flutter
#RUN chown -R flutter:flutter /flutter

# Ensure the home directory is owned by the flutter user
#RUN chown -R flutter:flutter /home/flutter

#ENV PATH="/flutter/bin:${PATH}"
ENV PATH="$SDK/bin:$SDK/bin/cache/dart-sdk/bin:${PATH}"

RUN mkdir $APP
COPY . $APP
WORKDIR $APP

# Switch to the non-root user
# USER flutter

# Create work directory and copy project
# WORKDIR /app
# COPY --chown=flutter:flutter . .

# Download dependencies and build
RUN flutter clean
RUN flutter pub get
RUN flutter build web
# removed:  --release --web-renderer html
RUN ls -l /app/build/web

# Stage 2: Create the Production Image with Nginx
FROM nginx:alpine

# Copy the built web app from the builder stage
# COPY --from=builder /app/build/web /usr/share/nginx/html
COPY --from=build-env /app/build/web /usr/share/nginx/html

# Expose port 8080
EXPOSE 80

# Start Nginx
CMD ["nginx", "-g", "daemon off;"]