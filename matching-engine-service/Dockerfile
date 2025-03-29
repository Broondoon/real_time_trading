FROM golang:1.23

WORKDIR /app

COPY matching-engine-service/ ./matching-engine-service
COPY Shared/ ./Shared
COPY stock-order-database/database-access ./databaseAccessStockOrder
COPY stock-database/database-access ./databaseAccessStock
COPY stock-order-database/database-service ./databaseServiceStockOrder
COPY user-management-database/database-access ./databaseAccessUserManagement

RUN go work init ./matching-engine-service
RUN go work use ./Shared
RUN go work use ./databaseAccessStockOrder
RUN go work use ./databaseServiceStockOrder
RUN go work use ./databaseAccessStock
RUN go work use ./databaseAccessUserManagement

WORKDIR /app/matching-engine-service

RUN go mod tidy 
RUN go build -o main .

CMD [ "./main"]
