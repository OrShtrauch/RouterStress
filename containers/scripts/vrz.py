import logging
import os
import signal
import sys
import time
from datetime import datetime
from logging import Logger

import requests
import data_handler

VRZ = "vrz"

name: str = os.environ["name"]
delay: int = int(os.environ["delay"])
dt_format: str = os.environ["dt_format"]

threads = int(os.environ["threads"])
concurrent = int(os.environ["concurrent"])

log_name: str = f"data/{name}.log"

logging.basicConfig(
    filename=log_name, format="%(asctime)s | %(name)s | %(message)s", level=logging.INFO
)
logger: Logger = logging.getLogger(VRZ)
logger.info("started running VRZ bug test")

data = list[tuple[str, str, int]]()
running: bool = True
last_request_time: float = None


urls = [
    "youtube.com",
    "facebook.com",
    "instagram.com",
    "twitch.com",
    "amazon.com"
]

def signal_handler(signal_number, frame):
    global running

    logger.info("got SIGTERM signal, calling data_handler")
    logger.info(f"sending data handler {len(data)} lines")
    running = False
    data_handler.run(data)
    logger.info("exiting...")
    sys.exit(0)


if __name__ == "__main__":
    signal.signal(signal.SIGTERM, signal_handler)

    def do():
        for _url in urls:
            response = None
            session = requests.Session()
            start: float = time.perf_counter()
            try:
                response = session.get(f"https://{_url}")
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
                        f"{{timestamp: {line[0]}, elapsed: {line[1]}, exit_code: {line[2]} url: }}"
                    )

                    time.sleep(delay)

            time.sleep(0.25)

    while running:
        do()