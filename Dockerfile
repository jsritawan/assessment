FROM golang:1.19-alpine as build-base
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go test --tags=unit -v ./...
RUN go build -o ./out/go-app .

FROM alpine
COPY --from=build-base /app/out/go-app /app/go-app
ARG DATABASE_URL
ARG PORT
ENV DATABASE_URL $DATABASE_URL
ENV PORT $PORT
EXPOSE $PORT
CMD ["/app/go-app"]