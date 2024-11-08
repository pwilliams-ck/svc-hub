FROM alpine:3.20

RUN mkdir /app

COPY duoApp /app

CMD [ "/app/duoApp" ]
