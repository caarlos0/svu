FROM alpine as build

RUN apk update && \
  apk add curl bash && \
  curl -sfL https://install.goreleaser.com/github.com/caarlos0/svu.sh | bash -s -- -b /usr/local/bin

FROM scratch

COPY --from=build /usr/local/bin/svu /usr/local/bin/svu

ENTRYPOINT ["svu"]
