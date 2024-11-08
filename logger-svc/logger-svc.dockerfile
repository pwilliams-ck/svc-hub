FROM alpine:3.20

RUN mkdir /app

COPY loggerApp /app

CMD [ "/app/loggerApp" ]
