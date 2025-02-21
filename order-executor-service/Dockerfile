FROM golang:1.23

WORKDIR /app

COPY order-executor-service/ ./OrderExecutorService
COPY Shared/ ./Shared
COPY transaction-database/database-access ./databaseAccessTransaction
COPY user-management-database/database-access ./databaseAccessUserManagement

RUN go work init ./OrderExecutorService
RUN go work use ./Shared
RUN go work use ./databaseAccessTransaction
RUN go work use ./databaseAccessUserManagement

WORKDIR /app/OrderExecutorService

RUN go mod tidy 
RUN go build -o main .

CMD [ "./main"]
