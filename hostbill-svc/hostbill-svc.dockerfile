FROM alpine:3.20

RUN mkdir /app

COPY hostbillApp /app

CMD [ "/app/hostbillApp" ]
