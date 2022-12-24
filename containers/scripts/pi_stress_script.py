import argparse
import os
import sys
import time
import requests
import logging
import signal

import dns.resolver as dns_res
import scapy.all as scapy

from datetime import datetime
from time import sleep

FILE_PATH = "data/network_data.csv"
DPI = "automation_dpi_header"
ARGS = None


def parse_args():
    # main parser, for arguments that are for all modes
    main_parser = argparse.ArgumentParser()
    main_parser.add_argument("-r", "--resume", action='store_true', help="overwrite existing data", default=False)
    sub_parser = main_parser.add_subparsers(dest="mode")

    http_parser = sub_parser.add_parser("http", help="flood a server with http get requests")
    http_parser.add_argument("-u", "--url", type=str, required=True, help="website url")
    http_parser.add_argument("-s", "--https", action='store_true', help="flag to use https")
    http_parser.add_argument("--dpi", action="store_true", default=False, help="use this to trigger a dpi event")

    dns_parser = sub_parser.add_parser("dns", help="flood with dns queries")
    dns_parser.add_argument("-d", "--domain", type=str, required=True, help="domain to resolve")
    dns_parser.add_argument("-t", "--type", type=str, default="A", help="record type: A, AAAA...etc")

    port_parser = sub_parser.add_parser("port", help="repeatedly scan top 100 ports on target")
    port_parser.add_argument("-t", "--target", type=str, required=True, help="ip of target")

    icmp_parser = sub_parser.add_parser("icmp", help="icmp flood the target")
    icmp_parser.add_argument("-t", "--target", type=str, required=True, help="ip of target")

    return main_parser.parse_args()


def http(url, scheme, dpi, file_stream, sleep_time=0):
    while True:
        response = None
        headers = {"User-Agent": DPI} if dpi else {}
        try:
            response = requests.get("%s://%s" % ("https" if scheme else "http", url), headers=headers)
        except Exception as e:
            logging.exception(e)
        finally:
            if response:
                line = "%s,%s,%s\n" % (datetime.now().strftime('%m/%d/%Y-%H:%M:%S.%f'), response.elapsed.total_seconds() * 1000,
                               0 if response.status_code == 200 else 1)
                logging.info(line)
                file_stream.writelines(line)

        sleep(sleep_time)


def dns(domain, record_type, file_stream, sleep_time=0):
    while True:
        result = None
        start = time.perf_counter()

        try:
            result = dns_res.query(domain, record_type)
        except Exception as e:
            logging.exception(e)
        finally:
            elapsed = round(time.perf_counter() - start, 5)

            if result:
                ip = result.response.answer[0].to_text().split()[-1]
                line = "%s,%s,%s\n" % (datetime.now().strftime('%m/%d/%Y-%H:%M:%S.%f'), elapsed, 0 if ip else 1)
            else:
                line = "%s,%s,%s\n" % (datetime.now().strftime('%m/%d/%Y-%H:%M:%S.%f'), elapsed, 1)

            logging.info(line)
            file_stream.writelines(line)

        sleep(sleep_time)


def portscan(target, file_stream, sleep_time=3):
    while True:
        nmap_cmd = "nmap -F --top-ports 100 %s" % target
        start = time.perf_counter()
        result = 0

        try:
            os.system(nmap_cmd)
        except Exception as e:
            logging.exception(e)
            result = 1
        finally:
            elapsed = round(time.perf_counter() - start, 5)

            line = "%s,%s,%s\n" % (datetime.now().strftime('%m/%d/%Y-%H:%M:%S.%f'), elapsed, result)

            logging.info(line)
            file_stream.writelines(line)

        sleep(sleep_time)


def icmp_flood(target, file_stream):
    result = 0

    while True:
        start = time.perf_counter()
        try:
            scapy.send(scapy.IP(dst=target) / scapy.ICMP())
        except Exception as e:
            logging.exception(e)
            result = 1
        finally:
            elapsed = round(time.perf_counter() - start, 5)
            line = "%s,%s,%s\n" % (datetime.now().strftime('%m/%d/%Y-%H:%M:%S.%f'), elapsed,
                                   0 if result == 0 else 1)
            file_stream.writelines(line)

            logging.info(line)


def handler(signum, frame):
    logging.info("Signal handler called at - %s - with signal: %s" % (datetime.now().strftime('%m/%d/%Y-%H:%M:%S.%f'),
                 signum))
    sys.exit(0)


if __name__ == '__main__':
    ARGS = parse_args()
    signal.signal(signal.SIGINT, handler)
    logging.basicConfig(filename="data/network_data.log")
    try:
        with open(FILE_PATH, "a+" if ARGS.resume else "w+") as file:
            if not ARGS.resume:
                file.writelines("timestamp,elapsed,status\n") # add test-number,sam-id,ssid to router and slaves
            # http("ynet.co.il", True, False, file)
            if ARGS.mode == "http":
                logging.info("%s:starting http flood" % (datetime.now().strftime('%m/%d/%Y-%H:%M:%S.%f')))
                http(ARGS.url, ARGS.https, ARGS.dpi, file)
            if ARGS.mode == "dns":
                logging.info("%s:starting dns flood" % (datetime.now().strftime('%m/%d/%Y-%H:%M:%S.%f')))
                dns(ARGS.domain, ARGS.type, file)
            if ARGS.mode == "port":
                logging.info("%s:starting port scan on %s" % (datetime.now().strftime('%m/%d/%Y-%H:%M:%S.%f'),
                                                                 ARGS.target))
                portscan(ARGS.target, file)
            if ARGS.mode == "icmp":
                logging.info("%s:starting icmp flood" % (datetime.now().strftime('%m/%d/%Y-%H:%M:%S.%f')))
                icmp_flood(ARGS.target, file)
    except Exception as e:
        logging.error(e)
        raise e
