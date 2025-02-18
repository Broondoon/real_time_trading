FROM golang:1.23

WORKDIR /app

COPY matching-engine-service/ ./matching-engine-service
COPY Shared/ ./Shared
COPY stock-order-database/database-access ./databaseAccessStockOrder
#Below is only present for dev. In prod, we will use a separate container for the database connection.
COPY stock-order-database/database-service ./databaseServiceStockOrder
COPY stock-transaction-database/database-access ./databaseAccessStockTransaction
COPY stock-transaction-database/database-service ./databaseServiceStockTransaction

RUN go work init ./matching-engine-service
RUN go work use ./Shared
RUN go work use ./databaseAccessStockOrder
RUN go work use ./databaseServiceStockOrder
RUN go work use ./databaseAccessStockTransaction
RUN go work use ./databaseServiceStockTransaction

WORKDIR /app/matching-engine-service

RUN go mod tidy 
RUN go build -o main .

CMD [ "./main"]
