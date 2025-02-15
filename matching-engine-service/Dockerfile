FROM golang:1.23

WORKDIR /app

COPY matching-engine-service/ ./matching-engine-service
COPY Shared/ ./Shared
COPY stock-order-database/database-access ./stock-order-database/database-access
#Below is only present for dev. In prod, we will use a separate container for the database connection.
COPY stock-order-database/database-service ./stock-order-database/database-service

RUN go work init ./matching-engine-service
RUN go work use ./Shared

WORKDIR /app/matching-engine-service

RUN go mod tidy 
RUN go build -o main .

CMD [ "./main"]
