## Use Go base image
#FROM golang:1.23
#
## Set working directory inside the container
#WORKDIR /app
#
## Copy auth-service and auth-database directories into the container.
#COPY auth-service/ ./auth-service
#COPY auth-database/ ./auth-database
#COPY Shared/ ./Shared
#
## Initialize Go workspace for the auth-service.
#RUN go work init ./auth-service
#RUN go work use ./Shared
#RUN go work use ./auth-database
#
## Move to the auth-service directory
#WORKDIR /app/auth-service
#
## Download dependencies and build the service
#RUN go mod tidy
#RUN go build -o auth-service .
#
## Command to run the service
#CMD ["./auth-service"]
FROM golang:1.23

WORKDIR /app

# Copy the necessary directories.
COPY auth-service/ ./auth-service
COPY auth-database/ ./auth-database
COPY Shared/ ./Shared
#COPY go.work go.work.sum ./

# Initialize Go workspace for auth-service.
# Instead of only using ./auth-database, add the specific subdirectories.
RUN go work init ./auth-service
RUN go work use ./Shared
RUN go work use ./auth-database/database-access
RUN go work use ./auth-database/database-service

WORKDIR /app/auth-service

RUN go mod tidy
RUN go build -o auth-service .

CMD [ "./auth-service" ]

