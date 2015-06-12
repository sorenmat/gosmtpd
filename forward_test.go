package main

import "testing"

func TestForwarding(t *testing.T) {
	/*
		FORWARD_ENABLED = true
		forwardhost = forwardHost()
		forwardport = forwardPort()
		// start the forward host
		config := MailConfig{hostname: "localhost", port: "2626", forwardEnabled: true, forwardHost: "localhost", forwardPort: "2626"}

		os.Setenv("PORT", "8181")
		go serve(config)

		SendMail("tester@localhost")
		FORWARD_ENABLED = false
		fmt.Println(config.database)
	*/
}

func forwardHost() *string {
	test := "localhost"
	return &test
}

func forwardPort() *string {
	test := "2525"
	return &test
}
