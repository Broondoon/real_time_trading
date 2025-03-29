FROM golang:1.23

WORKDIR /app

COPY order-executor-service/ ./OrderExecutorService
COPY Shared/ ./Shared
COPY user-management-database/database-access ./databaseAccessUserManagement

RUN go work init ./OrderExecutorService
RUN go work use ./Shared
RUN go work use ./databaseAccessUserManagement

WORKDIR /app/OrderExecutorService

RUN go mod tidy 
RUN go build -o main .

CMD [ "./main"]
