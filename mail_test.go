package main

import "testing"

func TestCleanUpEmail(t *testing.T) {
	email := "<sorenm@mymessages.dk>"
	ce, err := cleanupEmail(email)
	if err != nil {
		t.Error(err)
	}
	if ce != "sorenm@mymessages.dk" {
		t.Error("Address is wrong, when cleaning emails")
	}
}
func TestIsEmailAddressesValid(t *testing.T) {
	/*	mc := MailConnection{}
		mc.From = "test@test.com"
		mc.To = "to@test.com"

		scanForSubject(&mc, "subject: testing")
		err := isEmailAddressesValid(&mc)
		if err != nil {
			t.Error(err)
		}
		if mc.To != "to@test.com" {
			t.Error("To is wrong")
		}
		if mc.CC != "joe@blow.dk" {
			t.Error("CC is wrong")
		}
		if mc.Bcc != "test123@test.com" {
			t.Error("Bcc is wrong")
		}

		if mc.From != "test@test.com" {
			t.Error("From is wrong")
		}
	*/
}
