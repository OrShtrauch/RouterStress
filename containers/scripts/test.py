import signal
import sys
import time

if __name__ == '__main__':

    with open(f"data/file.test", "a+") as f:
        while True:
            f.write("reg")
            time.sleep(2)
