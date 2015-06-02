# A very simple SMTP server with a REST API

[![Travis CI](https://travis-ci.org/sorenmat/gosmtpd.svg?branch=master)](https://travis-ci.org/sorenmat/gosmtpd)
[![Coverage Status](https://coveralls.io/repos/sorenmat/gosmtpd/badge.svg)](https://coveralls.io/r/sorenmat/gosmtpd)

A server that accepts smtp request and saves the emails in memory for later retrieval.

/mail List all mails

/inbox/:email List all email for a given email address

/email/:id Get an email by id

