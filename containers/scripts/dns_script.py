import os
import logging
import data_handler
import time
import dns.resolver
import signal
import sys

from datetime import datetime
from concurrent.futures import ThreadPoolExecutor
from logging import Logger

DNS = "dns"

name: str = os.environ["name"]
delay: int = int(os.environ["delay"])
domain: str = os.environ["domain"]
record_type: str = os.environ["record_type"]
resolver: str = os.environ["resolver"]
dt_format: str = os.environ["dt_format"]

threads = int(os.environ["threads"])
concurrent = int(os.environ["concurrent"])

log_name: str = f"data/{name}.log"

logging.basicConfig(
    filename=log_name, format=f"%(asctime)s | %(name)s | %(message)s", level=logging.INFO
)
logger: Logger = logging.getLogger(DNS)
logger.info("started running dns flood")


data = list[tuple[str, str, int]]()
last_request_time: float = None
running: bool = True


def signal_handler(signal_number, frame):
    global running

    logger.info("got SIGINT signal, calling data_handler")
    logger.info(f"sending data handler {len(data)} lines")
    running = False
    data_handler.run(data, domain=domain, record_type=record_type)
    logger.info("exiting...")
    sys.exit(0)


if __name__ == "__main__":
    signal.signal(signal.SIGINT, signal_handler)

    dns_res = dns.resolver.Resolver()
    dns_cache = dns.resolver.Cache()

    if resolver != "default":
        dns_res.nameservers = [resolver]

    def do():
        response = None
        start = time.perf_counter()

        try:
            response = dns_res.resolve(qname=domain, rdtype=record_type)
        except Exception as e:
            logger.info("an error occurred when sending the request")
            logger.info(str(e))
        finally:
            elapsed: str = "{:.3f}".format(round(time.perf_counter() - start, 5))
            last_request_time = time.perf_counter()

            now: str = datetime.now().strftime(dt_format)
            exit_code: int = 0

            if response:
                try:
                    ip = response.response.answer[0].to_text().split()[-1]
                except IndexError:
                    exit_code = 1
            else:
                exit_code = 1

            line: tuple[str, str, int] = now, elapsed, exit_code

            data.append(line)

            logger.info(
                f"{{timestamp: {line[0]}, elapsed: {line[1]}, exit_code: {line[2]}}}"
            )

            time.sleep(delay)

            dns_cache.flush()

    while running:
        with ThreadPoolExecutor(max_workers=threads) as exec:
            [exec.submit(do) for i in range(concurrent)]
