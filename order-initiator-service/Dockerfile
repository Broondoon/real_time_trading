FROM golang:1.23

WORKDIR /app

COPY order-initiator-service/ ./OrderInitiatorService
COPY Shared/ ./Shared
COPY transaction-database/database-access ./databaseAccessTransaction

RUN go work init ./OrderInitiatorService
RUN go work use ./Shared
RUN go work use ./databaseAccessTransaction

WORKDIR /app/OrderInitiatorService

RUN go mod tidy 
RUN go build -o main .

CMD [ "./main"]
