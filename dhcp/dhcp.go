// Package responsible for implementing a dhcp server
package dhcp

import (
	"RouterStress/consts"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

type EmptyPool struct {
	msg string
}

func (e *EmptyPool) Error() string {
	return "No Available IP address For Lease"
}

type DHCPServer struct {
	NetworkId string
	Pool      []string
	Used      []string
}

func NewDHCPServer(networkId string) *DHCPServer {
	server := &DHCPServer{
		NetworkId: networkId,
		Pool:      make([]string, 0),
		Used:      make([]string, 0),
	}

	server.PopulatePool()

	return server
}

func (server *DHCPServer) PopulatePool() {
	subnetIp := strings.Join(strings.Split(server.NetworkId, ".")[:3], ".")

	var wg sync.WaitGroup
	for i := 1; i < 255; i++ {
		wg.Add(1)
		ip := fmt.Sprintf("%v.%v", subnetIp, i)

		go func() {
			cmd := exec.Command(consts.PING, "-c 4", "-w 5", ip)
			_, err := cmd.Output()

			// err != nil, when status code is 1, meaning the ping failed
			if err != nil {
				server.Pool = append(server.Pool, ip)
			}

			defer wg.Done()
		}()
	}

	wg.Wait()
}

func (server *DHCPServer) Lease() (string, error) {
	var err error

	if len(server.Pool) > 0 {
		ip := server.Pool[len(server.Pool)-1]
		server.Pool = server.Pool[:len(server.Pool)-1]
		
		server.Used = append(server.Used, ip)

		return ip, err
	} else {
		return "", &EmptyPool{msg: "No Available IP address For Lease"}
	}
}

func (server *DHCPServer) Release(ip string) {
	server.Pool = append(server.Pool, ip)
	server.Used = remove(server.Used, ip)
}

func remove(slice []string, ip string) []string {
	index := search(slice, ip)

	if index != -1 {
		return append(slice[:index], slice[index+1:]...)
	} else {
		return slice
	}
}

func search(slice []string, ip string) int {
	for i := 0; i < len(slice); i++ {
		if slice[i] == ip {
			return i
		}
	}
	return -1
}
