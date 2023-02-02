package consts

import (
	"github.com/google/uuid"
)

var (
	TEST_UUID = uuid.New().String()
	TEST_ID   = ""
	RUN_INDEX = 0

	TRAFFIC_CAPTURE_URL = "http://testmymalwarefiles.com/CH7465LG-NCIP-6.15.32p4-SH.p7"
)

const (
	TCP         = "tcp"
	BUFFER_SIZE = 8192
	PROMPT      = "#sam_prompt#"
	SSH         = "ssh"
	TELNET      = "telnet"
	SOCKET_PORT = "8192"
	NETCAT      = "nc"
	AWK         = "awk"
	PING        = "ping"
	MACVLAN     = "macvlan"
	SIGTERM     = 15
	UNIX        = "unix"

	INITIAL_CAPTURE_DURATION = 60
	RT_MAX_RATE              = 0.95

	BUCKET = "agent-stress-test-results-dev"

	THREADS        = 25
	CONCURRENT     = 25
	DT_FORMAT      = "%Y-%m-%d-%H:%M:%S"
	DT_LAYOUT      = "2006-01-02-15:04:05"
	PLOT_DT_LAYOUT = "15:04"
	TZ             = "Asia/Jerusalem"
	DELAY          = 1

	RESULTS_DIR        = "results/"
	SAMPLER_NAME       = "router_sampler.sh"
	SAMPLER_PATH       = "/var/tmp/" + SAMPLER_NAME
	SAMPLER_DATA_PATH  = "/var/tmp/hardware_data.csv"
	SAMPLER_LOCAL_NAME = "router_data.csv"
	TRAFFIC_DATA_NAME  = "packet_loss.json"

	TRAFFIC_UNIX_SOCKET = "/tmp/traffic.sock"

	DOCKERFILES_PATH           = "containers/"
	CONTAINER_SCRIPTS          = DOCKERFILES_PATH + "scripts"
	TRAFFIC_CAPTURE_PATH       = DOCKERFILES_PATH + "traffic_capture/"
	TRAFFIC_CAPTURE_IMAGE_NAME = "traffic_capture"

	DATA_PATH          = "data/"
	ROUTERS_PATH       = DATA_PATH + "routers.json"
	CONF_PATH          = DATA_PATH + "conf.json"
	SCENARIOS_PATH     = DATA_PATH + "scenarios.json"
	LOCAL_SAMPLER_PATH = DATA_PATH + "router_sampler.sh"

	STRESS_CONTAINER_PREFIX  = "stress"
	TRAFFIC_CONTAINER_PREFIX = "traffic_capture"
	CONTAINER_VERSION        = "latest"

	REMOTE_VOLUME_PATH = "/var/tmp/stress/data"
)
