FROM alpine:3.19
RUN apk add --no-cache curl bash grep
WORKDIR /tests
COPY smoke_test.sh ./
RUN chmod +x smoke_test.sh
CMD ["./smoke_test.sh"]
