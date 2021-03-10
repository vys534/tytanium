FROM golang:1.15-alpine AS build
RUN apk update && apk add --no-cache git
RUN apk --no-cache add ca-certificates

WORKDIR /build/lh

COPY go.mod ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 go build -o /bin/lh

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /bin/lh /bin/lh
COPY --from=build /build/lh/conf /bin/conf

WORKDIR /bin
EXPOSE 3030
ENTRYPOINT ["lh"]