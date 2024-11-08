FROM alpine:3.20

RUN mkdir /app

COPY authApp /app

CMD [ "/app/authApp" ]
