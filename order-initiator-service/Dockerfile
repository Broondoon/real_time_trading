FROM golang:1.23

WORKDIR /app

COPY order-initiator-service/ ./OrderInitiatorService
COPY Shared/ ./Shared
COPY user-management-database/database-access ./databaseAccessUserManagement
COPY stock-database/database-access ./databaseAccessStock

RUN go work init ./OrderInitiatorService
RUN go work use ./Shared
RUN go work use ./databaseAccessUserManagement
RUN go work use ./databaseAccessStock

WORKDIR /app/OrderInitiatorService

RUN go mod tidy 
RUN go build -o main .

CMD [ "./main"]
