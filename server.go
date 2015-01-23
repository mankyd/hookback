package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
)

type HookHandler struct {
	conf Config
}

type PayloadRepo struct {
	Full_Name string
}

type Payload struct {
	Repository PayloadRepo
}

// CheckMAC returns true if messageMAC is a valid HMAC tag for message.
func CheckMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha1.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}

func buildCmd(cmdStr []string, payload *string) *exec.Cmd {
	cmdArgs := cmdStr[1:]
	// Substitute in our payload wherever '%p' is found. The '%' character can be escaped with a
	// single, additional '%', so '%%p' becomes '%p', and '%%%p' becomes '%%p', etc.
	for argIndex, arg := range cmdArgs {
		placholderState := 0
		for _, c := range arg {
			if c == '%' && (placholderState == 0 || placholderState == 1) {
				placholderState = 1
			} else if c == 'p' && placholderState == 1 {
				placholderState = 2
			} else {
				placholderState = 3
				break
			}
		}
		if placholderState == 2 {
			if arg == "%p" {
				cmdArgs[argIndex] = *payload
			} else {
				cmdArgs[argIndex] = cmdArgs[argIndex][1:]
			}
		}
	}
	return exec.Command(cmdStr[0], cmdArgs...)
}

func runCmd(cmd *exec.Cmd, errChan chan error) {
	err := cmd.Start()
	if err != nil {
		errChan <- err
		return
	}
	// Command started successfully, so our caller can continue.
	errChan <- nil

	err = cmd.Wait()

	// Finally, tell our caller that we're complete, possibly with an error.
	errChan <- err
}

func getEventType(headers *http.Header) (string, error) {
	eventType := headers.Get("X-GitHub-Event")
	if eventType == "" {
		return eventType, errors.New("No event type set")
	}

	return eventType, nil
}

func getPayload(headers *http.Header, body *[]byte) (Payload, string, error) {
	var payload Payload
	var payloadS string

	contentType := headers.Get("content-type")
	if contentType == "application/x-www-form-urlencoded" {
		values, err := url.ParseQuery(string(*body))
		if err != nil {
			return payload, payloadS, err
		}
		pValues := values["payload"]
		if len(pValues) == 0 {
			return payload, payloadS, errors.New("No payload found")
		}
		payloadS = pValues[0]
	} else if contentType == "application/json" {
		payloadS = string(*body)
	} else {
		return payload, payloadS, errors.New("Unknown content type")
	}

	err := json.Unmarshal([]byte(payloadS), &payload)
	if err != nil {
		return payload, payloadS, err
	}

	return payload, payloadS, nil
}

func validateSecret(secret string, headers *http.Header, body *[]byte) error {
	if secret != "" {
		hmacHex := headers.Get("X-Hub-Signature")

		//TODO(mankyd): Support other hash types?
		if !strings.HasPrefix(hmacHex, "sha1=") {
			return fmt.Errorf("Unknown hash type: %s", hmacHex)
		}

		hmacSig, err := hex.DecodeString(hmacHex[5:])
		if err != nil {
			return err
		}

		if !CheckMAC(*body, hmacSig, []byte(secret)) {
			return errors.New("Invalid HMAC")
		}
	}

	return nil
}

func (hh *HookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// We need to read the body ourselves so that we can run an HMAC against it.
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}

	payload, payloadS, err := getPayload(&r.Header, &body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, err.Error())
		return
	}

	repoSettings, ok := hh.conf.Repos[payload.Repository.Full_Name]
	if !ok {
		errMsg := "Repository not found in config"
		log.Println(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, err.Error())
		return
	}

	err = validateSecret(repoSettings.Secret, &r.Header, &body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, err.Error())
		return
	}

	eventType, err := getEventType(&r.Header)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, err.Error())
		return
	}


	eventHandler, ok := repoSettings.Events[eventType]
	if !ok {
		w.WriteHeader(http.StatusPreconditionFailed)
		fmt.Fprintf(w, "Unhandled event type: %s", eventType)
		return
	}

	errChan := make(chan error)
	fmt.Println("running")
	cmd := buildCmd(eventHandler.Cmd, &payloadS)
	// If we're waiting for the command, capture its stdout/stderr.
	var outBuff bytes.Buffer
	if eventHandler.Wait.(bool) {
		cmd.Stdout = &outBuff
		cmd.Stderr = &outBuff
	} else {
		cmd.Stdout = nil
		cmd.Stderr = nil
	}
	go runCmd(cmd, errChan)
	err = <-errChan
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}

	if eventHandler.Wait.(bool) {
		err = <- errChan
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			state := int(cmd.ProcessState.Sys().(int))
			fmt.Fprintf(w, "Exit Status %d: %s\n", state, err)
		} else {
			fmt.Fprintf(w, "OK\n%s\n", eventHandler.Cmd)
			outBuff.WriteTo(w)
		}
	} else {
		// At this point, we're done with the connection. Flush, hijack, and close it.
		fmt.Fprintf(w, "OK\n%s", eventHandler.Cmd)
		
		fl, ok := w.(http.Flusher)
		if !ok {
			log.Println("WebServer does not support flushing.")
			return
		}
		fl.Flush()

		hj, ok := w.(http.Hijacker)
		if !ok {
			log.Println("WebServer does not suport hijacking.")
			return
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			log.Println("Problem hijacking the connection: %s", err)
			return
		}
		conn.Close()
		
		// Wait for the command to finish
		err = <- errChan
		if err != nil {
			log.Println("Problem completing command: %s", err)
		}
	}
	fmt.Println("complete")
}
