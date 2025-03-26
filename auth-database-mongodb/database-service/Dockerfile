FROM golang:1.23

WORKDIR /app

COPY auth-database-mongodb/database-service ./databaseServiceAuth
COPY Shared/ ./Shared

RUN go work init ./databaseServiceAuth
RUN go work use ./Shared

WORKDIR /app/databaseServiceAuth

RUN go mod tidy
RUN go build -o main .

CMD [ "./main"]

