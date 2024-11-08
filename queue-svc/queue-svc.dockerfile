FROM alpine:3.20

RUN mkdir /app

COPY queueApp /app

CMD [ "/app/queueApp" ]
