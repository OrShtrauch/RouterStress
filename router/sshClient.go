package router

import (
	"RouterStress/consts"
	"bytes"
	"fmt"

	"golang.org/x/crypto/ssh"
)

type SSHClient struct {
	Ip       string
	Port     string
	Username string
	Password string
	client   *ssh.Client
	session  *ssh.Session
}

func NewSSHClient(ip string, port string, username string, password string) (*SSHClient, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	host := fmt.Sprintf("%v:%v", ip, port)
	client, err := ssh.Dial(consts.TCP, host, config)

	if err != nil {
		return nil, err
	}

	return &SSHClient{
		Ip:       ip,
		Port:     port,
		Username: username,
		Password: password,
		client:   client,
	}, err
}

func (c *SSHClient) Run(cmd string) (string, error) {
	var err error

	if c.session != nil {
		c.session.Close()
	}

	session, err := c.client.NewSession()
	c.session = session
	
	if err != nil {
		fmt.Printf("session")
		return "", err
	}

	var b bytes.Buffer
	c.session.Stdout = &b

	err = c.session.Run(cmd)
	data := b.String()

	if err != nil {
		fmt.Println(err)
		return data, err
	}

	return data, err
}

func (c *SSHClient) CloseListenerSession(cmd string) {
	if c.session != nil {
		c.session.Close()
		c.session = nil
	}
}

func (c *SSHClient) Close() {
	c.client.Close()
}
