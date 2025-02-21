# Use the official Go image
FROM golang:1.23

WORKDIR /app

COPY user-management-tests/ ./user-management-tests
COPY Shared/ ./Shared
COPY user-management-service/ ./user-management-service

RUN go work init ./user-management-tests
RUN go work use ./Shared
RUN go work use ./user-management-service

WORKDIR /app/user-management-tests
RUN go mod tidy

CMD ["go", "test", "-v", "./..."]
