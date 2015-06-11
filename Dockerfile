FROM scratch
MAINTAINER Soren Mathiasen <sorenm@mymessages.dk>

ADD gosmtpd /
EXPOSE 2525
CMD ["/gosmtpd"]
