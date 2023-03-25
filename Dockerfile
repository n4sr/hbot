FROM alpine

ENV TOKEN=${TOKEN}

COPY ./*go* /
COPY ./run.sh /

RUN apk add golang
RUN go build
RUN chmod +x hbot

ENTRYPOINT ["/run.sh"]
