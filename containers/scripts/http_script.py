import logging
from operator import concat
import os
import signal
import sys
import time
from datetime import datetime
from concurrent.futures import ThreadPoolExecutor
from logging import Logger

import httpx
import data_handler

HTTP = "http"

url: str = os.environ["url"]
name: str = os.environ["name"]
delay: int = int(os.environ["delay"])
scheme: str = os.environ["scheme"].lower()
dt_format: str = os.environ["dt_format"]

threads = int(os.environ["threads"])
concurrent = int(os.environ["concurrent"])

log_name: str = f"data/{name}.log"

logging.basicConfig(
    filename=log_name, format="%(asctime)s | %(name)s | %(message)s", level=logging.INFO
)
logger: Logger = logging.getLogger(HTTP)
logger.info("started running http flood")

data = list[tuple[str, str, int]]()
running: bool = True
last_request_time: float = None

def signal_handler(signal_number, frame):
    global running

    logger.info("got SIGINT signal, calling data_handler")
    logger.info(f"sending data handler {len(data)} lines")
    running = False
    data_handler.run(data, url=url, scheme=scheme)
    logger.info("exiting...")
    sys.exit(0)


if __name__ == "__main__":
    signal.signal(signal.SIGINT, signal_handler)

    def do():
        response = None
        start: float = time.perf_counter()
        try:
            response = httpx.get(f"{scheme}://{url}", follow_redirects=True)
        except Exception as e:
            logger.info("an error occurred when sending the request")
            logger.info(str(e))
        finally:
            elapsed: str = "{:.3f}".format(round(time.perf_counter() - start, 5))
            last_request_time = time.perf_counter()
            if response:
                now: str = datetime.now().strftime(dt_format)
                exit_code: int = 0 if response.status_code / 100 == 2 else 1
                line: tuple[str, str, int] = now, elapsed, exit_code

                data.append(line)

                logger.info(
                    f"{{timestamp: {line[0]}, elapsed: {line[1]}, exit_code: {line[2]}}}"
                )

                time.sleep(delay)

    while running:
        with ThreadPoolExecutor(max_workers=threads) as exec:
            [exec.submit(do) for i in range(concurrent)]
