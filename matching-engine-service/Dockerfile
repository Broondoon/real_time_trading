FROM golang:1.23

WORKDIR /app

COPY matching-engine-service/ ./matching-engine-service
COPY Shared/ ./Shared

RUN go work init ./matching-engine-service
RUN go work use ./Shared

WORKDIR /app/matching-engine-service

RUN go mod tidy 
RUN go build -o main .

CMD [ "./main"]
