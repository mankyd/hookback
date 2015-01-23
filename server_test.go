package main

import (
	"bytes"
	"os/exec"
	"testing"
)

func TestCheckMAC(t *testing.T) {
	type testData struct {
		message, key, mac []byte
	}

	cases := []testData{
		{
			[]byte{'a', 'b', 'c'},
			[]byte{'d', 'e', 'f'},
			[]byte{117, 132, 238, 20, 73, 48, 114, 205, 141, 154, 232, 80, 224, 255, 9, 5, 56, 35, 15, 140},
		},
		testData{
			[]byte{0, 1, 2, 3},
			[]byte{4, 5, 6},
			[]byte{2, 46, 215, 172, 137, 54, 176, 174, 166, 161, 227, 241, 48, 52, 109, 222, 37, 234, 234, 108}},
	}

	for _, c := range cases {
		if !CheckMAC(c.message, c.mac, c.key) {
			t.Errorf("HMAC Does not pass. Message: %s. Key: %s. MAC: %#v.", c.message,c.key, c.mac)
		}
	}
}

func TestRunCmd(t *testing.T) {
	errChan := make(chan error)
	expectedOuput := "foobar"
	cmd := exec.Command("echo", "-n", expectedOuput)
	var out bytes.Buffer
	cmd.Stdout = &out

	go runCmd(cmd, errChan)

	err := <- errChan
	if err != nil {
		t.Errorf("Error starting command: %s", err)
		return
	}
	err = <- errChan
	if err != nil {
		t.Errorf("Error running command: %s", err)
		return
	}

	if out.String() != expectedOuput {
		t.Errorf("Unexepected command out. Expected \"%s\", got \"%s\".", expectedOuput, out.String())
	}
}
