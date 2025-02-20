# Use Go base image
FROM golang:1.23

# Set working directory inside the container
WORKDIR /app

# Copy source code into the container
COPY user-management-service/ ./user-management-service
COPY Shared/ ./Shared
COPY user-management-database/database-access ./databaseAccessUserManagement
COPY user-management-database/database-service ./databaseServiceUserManagement

# Initialize Go workspace
RUN go work init ./user-management-service
RUN go work use ./Shared
RUN go work use ./databaseAccessUserManagement
RUN go work use ./databaseServiceUserManagement

# Move to the user-management-service directory
WORKDIR /app/user-management-service

# Download dependencies and build the service
RUN go mod tidy
RUN go build -o user-management-service .

# Expose the service port (update if needed)
EXPOSE 8080

# Command to run the service
CMD ["./user-management-service"]
