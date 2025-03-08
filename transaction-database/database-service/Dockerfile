FROM golang:1.23

WORKDIR /app

COPY transaction-database/database-service ./databaseServiceTransaction
COPY Shared/ ./Shared

RUN go work init ./databaseServiceTransaction
RUN go work use ./Shared

WORKDIR /app/databaseServiceTransaction

RUN go mod tidy 
RUN go build -o main .

CMD [ "./main"]
