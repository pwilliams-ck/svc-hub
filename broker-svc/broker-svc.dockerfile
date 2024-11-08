FROM alpine:3.20

RUN mkdir /app

COPY brokerApp /app

CMD [ "/app/brokerApp" ]
