package consts

import (
	"fmt"

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
	MACVLAN     = "macvlan"
	SIGTERM     = 15

	THREADS    = 25
	CONCURRENT = 25
	DT_FORMAT  = "%Y-%m-%d-%H:%M:%S"
	TZ         = "Asia/Jerusalem"
	DELAY      = 1

	SAMPLER_NAME      = "router_sampler.sh"
	SAMPLER_PATH      = "/var/tmp/" + SAMPLER_NAME
	SAMPLER_DATA_PATH = "/var/tmp/hardware_data.csv"

	DOCKERFILES_PATH  = "./containers"
	CONTAINER_SCRIPTS = DOCKERFILES_PATH + "/scripts"
)

var (
	ROUTERS_PATH       = "data/routers.json"
	CONF_PATH          = "data/conf.json"
	SCENARIOS_PATH     = "data/scenarios.json"
	LOCAL_SAMPLER_PATH = "data/router_sampler.sh"
	TRAFFIC_DATA_PATH  = "data/traffic_data.json"

	CONTAINER_VERSION = "latest"
	TestUUID          = uuid.New().String()
	TestID            = ""
	Run_index         = 0

	LOCAL_VOLUME_PATH  = fmt.Sprintf("results/%v", TestID)
	REMOTE_VOLUME_PATH = "/var/tmp/stress/data"
)
