
# StressLI
SAM internal tool for stress testing routers, using different custom made python scripts.

The tool's behavior is defined by json config file.
And it's scenarios are defined by a scenarios.json file

## Folder Structure
In order for the program to run successfully, you must keep the same structure (a tar.gz archive with the current structure can be found in the github repository)
```
StressLI/
-stress
-stress.log
-containers/
--scripts/
---*python_script.py*
--*Dockerfiles*
-data/
--conf.json
--router_sampler.sh
--routers.json
--scenarios.json
-results/
--*TEST_ID*/
```

To run the binary just execute
```
./stress&
```

Its important to keep the names of the files.

Any output will be directed to *stress.log* file, including error messages, and test progress