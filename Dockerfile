FROM golang
MAINTAINER Soren Mathiasen <sorenm@mymessages.dk>
ADD run.sh /
ADD . /go/src/github.com/sorenmat/gosmtpd
WORKDIR /go/src/github.com/sorenmat/gosmtpd
RUN go get -v
RUN go install github.com/sorenmat/gosmtpd
EXPOSE 2525
CMD ["/go/bin/gosmtpd"]
