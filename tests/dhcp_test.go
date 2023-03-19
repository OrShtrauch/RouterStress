package tests

import (
	"RouterStress/dhcp"
	"testing"
)

const (
	subnet = "10.0.0.0"
)

var server *dhcp.DHCPServer

// Test DHCP Server
// TestCreateDhcpServer - create server and check if created
// TestLeaseAndReleaseAddr - leases ip and makes sure its in the used pool, the releases it,
// and checks its no loger there, and back in the pool

func TestCreateDhcpServer(t *testing.T) {
	server = dhcp.NewDHCPServer(subnet)

	if server == nil || server.NetworkId != subnet {
		t.Error("Expected DHCP to be DHCPServer Obj.")
	}
}

func TestLeaseAndReleaseAddr(t *testing.T) {
	ip, err := server.Lease()

	if err != nil {
		t.Error("Pool should be full")
	}

	found := false

	for _, used_ip := range server.Used {
		if used_ip == ip {
			found = true
		}
	}

	if !found {
		t.Errorf("Expected %v to be in the used pool", ip)
	}

	server.Release(ip)

	found = false
	for _, used_ip := range server.Used {
		if used_ip == ip {
			found = true
		}
	}

	if found {
		t.Errorf("Expected %v not to be in the used pool", ip)
	}

	found = false
	for _, used_ip := range server.Pool {
		if used_ip == ip {
			found = true
		}
	}

	if !found {
		t.Errorf("Expected %v to be in the availbale pool", ip)
	}
}

func TestLeaseEmptyPool(t *testing.T) {
	var err error	

	server.Pool = []string{}
	server.Lease()

	if err != nil && err.Error() != "No Available IP address For Lease" {
		t.Errorf("Expected EmptyPool Error, Got %v insted", err)
	}	
}