import logging
import os
import signal
import sys
import time
from datetime import datetime
from logging import Logger
from concurrent.futures import ThreadPoolExecutor

import scapy.all as scapy
import data_handler

ICMP = "icmp"

name: str = os.environ["name"]
delay: int = int(os.environ["delay"])
target: str = os.environ["target"]
dt_format: str = os.environ["dt_format"]

threads = int(os.environ["threads"])
concurrent = int(os.environ["concurrent"])


log_name: str = f"data/{name}.log"

logging.basicConfig(
    filename=log_name, format="%(asctime)s | %(name)s | %(message)s", level=logging.INFO
)
logger: Logger = logging.getLogger(ICMP)
logger.info("started running icmp flood")

data = list[tuple[str, str, int]]()
running: bool = True
last_request_time: float = None

def signal_handler(signal_number, frame):
    global running

    logger.info("got SIGINT signal, calling data_handler")
    logger.info(f"sending data handler {len(data)} lines")
    running = False
    data_handler.run(data, target=target)
    logger.info("exiting...")
    sys.exit(0)


if __name__ == "__main__":
    signal.signal(signal.SIGINT, signal_handler)

    def do():
        global last_request_time

        exit_code: int = 0
        start: float = time.perf_counter()

        try:
            scapy.send(scapy.IP(dst=target) / scapy.ICMP(id=1, seq=1))
        except Exception as e:
            logger.info("an error occurred when sending the request")
            logger.info(str(e))

            exit_code = 1
        finally:
            elapsed: str = "{:.3f}".format(round(time.perf_counter() - start, 5))
            last_request_time = time.perf_counter()
            now: str = datetime.now().strftime(dt_format)
            line: tuple[str, str, int] = now, elapsed, exit_code

            data.append(line)

            logger.info(
                f"{{timestamp: {line[0]}, elapsed: {line[1]}, exit_code: {line[2]}}}"
            )

            time.sleep(delay)

    while running:
        with ThreadPoolExecutor(max_workers=threads) as exec:
            [exec.submit(do) for i in range(concurrent)]
