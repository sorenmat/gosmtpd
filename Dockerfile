FROM scratch
MAINTAINER Soren Mathiasen <sorenm@mymessages.dk>

ADD gosmtpd /
EXPOSE 2525
EXPOSE 8080
CMD ["/gosmtpd"]
