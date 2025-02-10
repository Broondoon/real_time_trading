FROM golang:1.23

WORKDIR /app

COPY microservice-template/ ./microservice-template
COPY Shared/ ./Shared

RUN go work init ./microservice-template
RUN go work use ./Shared

WORKDIR /app/microservice-template

RUN go mod tidy 
RUN go build -o main .

CMD [ "./main"]
