package consts

import (
	"github.com/google/uuid"
)

const (
	TCP         = "tcp"
	BUFFER_SIZE = 8192
	PROMPT      = "#sam_prompt#"
	SSH         = "ssh"
	TELNET      = "telnet"
	SOCKET_PORT = "8654"
	NETCAT      = "nc"
	AWK         = "awk"
	PING        = "ping"

	SAMPLER_PATH      = "/var/tmp/router_sampler.sh"
	SAMPLER_DATA_PATH = "/var/tmp/hardware_data.csv"

	DOCKERFILES_PATH  = "./containers"
	CONTAINER_SCRIPTS = DOCKERFILES_PATH + "/scripts"
)

var (
	ROUTERS_PATH       = "/data/routers.json"
	CONF_PATH          = "./data/conf.json"
	SCENARIOS_PATH     = "./data/scenarios.json"
	LOCAL_SAMPLER_PATH = "./data/router_sampler.sh"

	TestID    = uuid.New().String()
	Run_index = 0
)
