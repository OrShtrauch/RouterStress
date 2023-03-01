package router

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"RouterStress/consts"
)

type TelnetClient struct {
	Ip      string
	Port    string
	prompt  string
	session net.Conn
}

func NewTelnetClient(ip string, port string, login ...string) (*TelnetClient, error) {
	host := fmt.Sprintf("%v:%v", ip, port)
	session, err := net.Dial(consts.TCP, host)

	client := &TelnetClient{
		Ip:      ip,
		Port:    port,
		prompt:  "#",
		session: session,
	}

	
	if len(login) > 0 {
		client.login(login)
	} else {			
		client.receiveUntilPrompt()
	}

	client.setPrompt(consts.PROMPT)

	return client, err
}

func (client *TelnetClient) login(creds []string) error {
	var err error
	duration := 3

	client.receiveForDuration(duration)

	for _, cred := range creds {
		_, err = client.session.Write([]byte(fmt.Sprintf("%v\n", cred)))

		if err != nil {
			return err
		}

		_, err = client.receiveForDuration(duration)		
		
		if err != nil && !os.IsTimeout(err) {
			return err
		}
	}

	_, err = client.receiveForDuration(duration)

	return err
}

func (client *TelnetClient) Run(cmd string) (string, error) {
	cmd_w_enter := fmt.Sprintf("%v\n", cmd)
	_, err := client.session.Write([]byte(cmd_w_enter))

	data := client.receiveUntilPrompt()
	trimmedOutput := client.trimOutput(data, cmd)

	return string(trimmedOutput), err
}

func (client *TelnetClient) trimOutput(data string, cmd string) string {
	data = strings.Replace(data, client.prompt, "", -1)
	data = strings.Replace(data, cmd, "", -1)

	return data
}

func (client *TelnetClient) receiveForDuration(duration int) (string, error) {
	var stringBuilder strings.Builder

	err := client.session.SetReadDeadline(time.Now().Add(time.Second * time.Duration(duration)))

	if err != nil {
		return stringBuilder.String(), err
	}

	start_time := time.Now()

	for int(time.Since(start_time).Seconds()) < duration {
		buffer := make([]byte, consts.BUFFER_SIZE)
		mLen, err := client.session.Read(buffer)

		if err != nil {
			return stringBuilder.String(), err
		}

		stringBuilder.Write(buffer[:mLen])
	}

	return stringBuilder.String(), err
}

func (client *TelnetClient) receiveUntilPrompt(opts ...int) string {
	var stringBuilder strings.Builder
	var timeout int

	found_prompt := false

	if len(opts) > 0 {
		for _, _timeout := range opts {
			timeout = _timeout
		}
	} else {
		timeout = 10
	}

	err := client.session.SetReadDeadline(time.Now().Add(time.Second * time.Duration(timeout)))

	if err != nil {
		return err.Error()
	}

	start_time := time.Now()

	timeout_passed := int(time.Since(start_time).Seconds()) > timeout

	for !found_prompt && !timeout_passed {
		buffer := make([]byte, consts.BUFFER_SIZE)
		mLen, err := client.session.Read(buffer)

		data := buffer[:mLen]
		stringBuilder.Write(data)

		if err != nil {
			return err.Error()
		}

		found_prompt = strings.Contains(string(data), client.prompt)
		timeout_passed = int(time.Since(start_time).Seconds()) > timeout
	}

	return stringBuilder.String()
}

func (client *TelnetClient) setPrompt(prompt string) error {
	_, err := client.Run(fmt.Sprintf("export PS1='%s'", prompt))
	client.prompt = prompt

	return err
}

func (client *TelnetClient) CloseListenerSession(cmd string) {
	client.Run(fmt.Sprintf("kill -9  `%s`", cmd))
}

func (client *TelnetClient) Close() {
	client.session.Close()
}
