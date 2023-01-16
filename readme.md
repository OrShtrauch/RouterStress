
# StressLI
SAM internal tool for stress testing routers, using different custom made python scripts.

The tool's behavior is defined by json config file.
And it's scenarios are defined by a scenarios.json file
## Prerequisites:

- A testing router with shell access, containing the following binaries: (could be using BusyBox or ROM binaries)
` awk`,` nc`

- A ***Linux*** machine connected to the tested router via an ethernet cable
- Docker Daemon up and running on the linux machine


## Folder Structure
In order for the program to run successfully, you must keep the same structure (a tar.gz archive with the current structure can be found in the github repository)

```
.
├── containers/
├──── scripts/
├────── *python_script.py*
├──── *Dockerfiles*
├──── Dockerfile.template
├── data/
├──── conf.json
├──── router_sampler.sh
├──── routers.json
├──── scenarios.json
├── results/
├──── *TestID*/
├── RouterStress
├── stress.log
└── README.md
```
Before Running the binary you have to modify conf.json,
which is located in the 'data/' directory,
there you will spesicify the ssid of the connected router, and the interface its conneted on.

To list available Ethernet Interfaces run the command:
```
lshw -class network -shorttwork -short
```

### More Information on the configuration file in the end

To run the binary just execute
```
./RouterStress &
```

Its important to keep the names of the files.

Any output will be directed to *stress.log* file, including error messages, and test progress

## Test Resutls
A test id will be assigned to each test, with the format of the ssid with a uuid (ssid_uuid),
the test id will be written to the log file

The test results will be in the results folder in a sub directory with the name of test id.
the results will consist of the following:

 - router_data.csv, csv file with cpu and memory data on the router while the test is running
 - stress_mode_*.csv, file spesifing each request a container make, the time it took, and it's exit code
 - TODO: cpu over time graph
 - TODO: aggregated json file

 ## Adding More Routers
 you can add new routers in 2 ways:
  1. use 'router_extractor.py' in the data directory, this will use the automation team 'routers.yaml' and 'firmwares.py' to generate json file with all the releavent fields.
  2. Add one manually

 ## Adding Custom Modes

 To add a custom mode you have to add it in the scenarios.json file, located in the 'data' directory
 there you will specify:
 - Name
 - Display Name
 - pip Dependencies
 - Script Path
 - Parameters To Pass

after adding it, a docker image will be created on runtime, so you can use the mode in the conf.json

## Config file
The config file is composed of 2 main parts, 

### Network 
network and router data (by default, only the ssid, and interface are needed)

#### Example:
```json
"network":
{
	"ssid": "Josh_HT138",
	"gateway": "10.0.0.138",
	"interface": "eno1",
	"network_id": "10.0.0.0/24"
}
```
 | field | value | 
 | ----- | ----- |
 | `ssid` | router's ssid |
 | `gateway` | router's default gateway ip address |
 | `interface` | parent interface of the host machine(must be connected to the router by an Ethernet cable) |
 | `network_id` | calculated from the gateway field (assuming the subnet is 24), for a custom one, specify it yourself manually in the config file |

### Iterations:
iteration is the main flow event, each iteration is defined by its duration, the protocols it uses,
and the cooldown time before the next iteration.
    
    
#### Example:
```json
 {
    "duration": "4",
    "cooldown": "10",
    "protocols":
    [
        {
            "mode": "HTTP",
            "containers":
            [                    
                {
                    "amount": "5",
                    "params":
                    {
                        "scheme": "HTTP",
                        "url": "10.10.1.11"
                    }
                }
            ]
        }
	]
}
```

| field | value |
| ----- | ----- | 
| `duration` |  length (in minuets) of the iterations |
| `cooldown` | length (in seconds) to wait before running the next iteration | 
| `protocols` | list of containers to run |

 - the protocol object contains a list of of containers to run for a given protocol, each item in containers, represent different params given to the container, for Example in HTTP, send requests to multiple websites at the same time


protocols fields:
```json
{
    "mode": "HTTP",
    "containers": [...]
}
```
| field | value |
| ----- | ----- | 
| `mode` | what network protocols(script) to run in the container |
| `containers` | a list specifying amount per intensity level, and param |

containers fields: 
```json
 {
    "amount": "5",
    "params": {...}
}
```
| field | value |
| ----- | ----- | 
| `amount` | amount of containers to run | 
| `params` | special params to the containers, per protocol. see params table below |

params example:
```json
"params":
	{
		"scheme": "HTTPS",
		"url": "10.10.1.11"
	}
```

### params table:
#### HTTP:
| field | value |
| ----- | ----- | 
| `url` | target url |
| `scheme`| HTTP/HTTPS |

#### DNS:
| field | value |
| ----- | ----- | 
| `domain` | domain to resolve |
| `record type`| A/AAAA/CNAME |
| `resolver`| dns server ip |


#### Example config file:
```json
{
    "network":
    {
        "ssid": "Josh_HT138",
        "gateway": "10.0.0.138",
        "interface": "eno1",
        "network_id": "10.0.0.0/24"
    },
    "iterations":
    [
        {
            "duration": "4",
            "cooldown": "10",
            "protocols":
            [
                {
                    "mode": "HTTP",
                    "containers":
                    [                    
                        {
                            "amount": "5",
                            "params":
                            {
                                "scheme": "HTTP",
                                "url": "10.10.1.11"
                            }
                        }
                    ]
                },
                {
                    "mode": "Port Scanning",
                    "containers":
                    [
                        {
                            "amount": "3",
                            "params":
                            {...}
                        }
                    ]
                }
            ]
        },
        {
            "duration": "7",
            "cooldown": "10",
            "protocols":
            [               
                {
                    "mode": "Brute Force",
                    "containers":
                    [
                        {
                            "amount": "3",
                            "params":
                            {...}
                        }
                    ]
                }
            ]
        }
    ]
}
```