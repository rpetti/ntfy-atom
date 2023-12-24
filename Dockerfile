FROM golang:1.21 AS build

COPY . /src

WORKDIR /src
RUN GCO_ENABLED=0 GOOS=linux go build -o /src/ntfy-atom

FROM debian:12-slim

RUN apt-get update && apt-get install -y ca-certificates && apt-get autoclean

COPY --from=build /src/ntfy-atom /ntfy-atom

USER 1000

EXPOSE 8080

CMD ["/ntfy-atom"]

