FROM --platform=linux/amd64 debian:stable-slim

RUN apt-get update && apt-get install -y ca-certificates

ADD phamily-photos /usr/bin/phamily-photos

CMD ["phamily-photos"]
