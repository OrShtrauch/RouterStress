import sys
import logging
import json
import os
import asyncio
import aiohttp
import signal

from logging import Logger
from aiohttp import ClientSession

# TODO: add better test_id = ssid + uuid
# create json for each entry
SSID: str = "ssid"
PLATFORM: str = "platform"
MODE: str = "mode"
TEST_ID: str = "test_id"
ITERATION: str = "iteration"
SENT_DATA: bool = False

test_id: str = os.environ[TEST_ID]

name: str = os.environ["name"]
log_name: str = f"{name}.log"
file_name: str = f"{name}.csv"
#path: str = f"data/{test_id}"
path: str = "data/"

logging.basicConfig(
    filename=f"{path}/{log_name}",
    format=f"%(asctime)s | %(name)s | %(message)s",
    level=logging.INFO,
)    

logger: Logger = logging.getLogger("data handler")

# ignore any sigint signal going to the script,
# the script will exit after sending data
signal.signal(signal.SIGTERM, signal.SIG_IGN)


def build_json(line: tuple[str, str, int], json_template: dict) -> str | None:
    try:
        json_template["time_stamp"] = line[0]
        json_template["elapsed"] = float(line[1])
        json_template["status"] = line[2]

        return json.dumps(json_template)
    except IndexError:
        return None


async def send_data(
    session: ClientSession, url: str, json_data: str, index: int
) -> None:
    try:
        logger.info(
            f"request #{index}"
        )  # , running post to: {url}, with body -> {json_data}")
        # await session.post(url=url, data=json_data)
        await asyncio.sleep(1)  # mock server request
    except Exception as e:
        logger.info(f"error at index: {index}, {str(e)}")


def handle_data(
    session: ClientSession, data: list[tuple[str, str, int]], **kwargs
) -> list:
    try:
        json_template: dict = {
            "ssid": os.environ[SSID],
            "platform": os.environ[PLATFORM],
            "mode": os.environ[MODE],
            "test_id": os.environ[TEST_ID],
            "iteration": os.environ[ITERATION],
            **{key: val for key, val in kwargs.items()},
        }
    except KeyError as e:
        logger.info(f"invalid environment variables | {str(e)}")
        sys.exit(0)

    tasks: list = []
    url: str = "http://10.0.0.11:8000"
    for index, line in enumerate(data):
        json_data: str = build_json(line, json_template)

        if json_data:
            tasks.append(send_data(session, url, json_data, index))

    return tasks


async def main(data: list[tuple[str, str, int]], **kwargs) -> None:
    global SENT_DATA

    logger.info("running cleanup")
    SENT_DATA = True
    async with aiohttp.ClientSession() as session:
        logger.info("started aiohtto session")
        tasks: list = handle_data(session, data, **kwargs)

        await asyncio.gather(*tasks)

        logger.info(f"sent {len(data)} lines of data")


def run(data: list[tuple[str, str, int]], **kwargs) -> None: 
    logger.info(f"got {len(data)} lines")
    logger.info(f"SENT -> {SENT_DATA}")
    
    with open(f"{path}/{file_name}", "w") as file:
        file.write("timestamp,elapsed,status\n")
        for line in data:
            file.writelines(f"{line[0]},{line[1]},{line[2]}\n")

    if not SENT_DATA:
        asyncio.run(main(data, **kwargs))
