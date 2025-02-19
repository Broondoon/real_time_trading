FROM golang:1.23

WORKDIR /app

COPY stock-database/database-service ./databaseServiceStock
COPY Shared/ ./Shared

RUN go work init ./databaseServiceStock
RUN go work use ./Shared

WORKDIR /app/databaseServiceStock

RUN go mod tidy 
RUN go build -o main .

CMD [ "./main"]
