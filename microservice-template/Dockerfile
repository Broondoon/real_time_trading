FROM golang:1.23

WORKDIR /app

COPY microservice-template/ ./microservice-template
COPY Shared/ ./Shared
# COPY stock-order-database/database-access ./databaseAccessStockOrder
# COPY transaction-database/database-access ./databaseAccessTransaction
# COPY user-management-database/database-access ./databaseAccessUserManagement

RUN go work init ./microservice-template
RUN go work use ./Shared
# RUN go work use ./databaseAccessStockOrder
# RUN go work use ./databaseAccessTransaction
# RUN go work use ./databaseAccessUserManagement

WORKDIR /app/microservice-template

RUN go mod tidy 
RUN go build -o main .

CMD [ "./main"]
