FROM golang:1.23

WORKDIR /app

COPY microservice-template/ ./microservice-template
COPY Shared/ ./Shared
COPY stock-order-database/database-access ./databaseAccessStockOrder
#Below is only present for dev. In prod, we will use a separate container for the database connection.
COPY stock-order-database/database-service ./databaseServiceStockOrder

RUN go work init ./microservice-template
RUN go work use ./Shared
RUN go work use ./databaseAccessStockOrder
RUN go work use ./databaseServiceStockOrder

WORKDIR /app/microservice-template

RUN go mod tidy 
RUN go build -o main .

CMD [ "./main"]
