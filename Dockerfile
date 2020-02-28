FROM golang:1.14
COPY . /pigeon
WORKDIR /pigeon
RUN go get github.com/cespare/reflex
RUN go get -u -a -v -x github.com/ipsn/go-libtor
EXPOSE 80
ENTRYPOINT ["reflex", "-c", "reflex.conf"]