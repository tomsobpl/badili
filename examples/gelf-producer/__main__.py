# Copyright 2026 Tomasz Sobczak <ictslabs@gmail.com>
# SPDX-License-Identifier: Apache-2.0

import logging
import traceback
import time
import os
import random
import signal
from pygelf import GelfUdpHandler


RUNNING = True

GELF_APP_NAME = os.environ.get("GELF_APP_NAME", "gelf-producer")

GELF_LOG_LEVEL = getattr(
    logging, os.environ.get("GELF_LOG_LEVEL", "DEBUG").upper(), logging.DEBUG
)

GELF_SLEEP_PERIOD_MS = int(os.environ.get("GELF_SLEEP_PERIOD_MS", 500))

GELF_TARGET_HOST = os.environ.get("GELF_TARGET_HOST", "127.0.0.1")
GELF_TARGET_PORT = int(os.environ.get("GELF_TARGET_PORT", 12201))


def signal_handler(signum, frame):
    global RUNNING
    RUNNING = False


def main():
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    handler = GelfUdpHandler(
        host=GELF_TARGET_HOST,
        port=GELF_TARGET_PORT,
        debug=True,
        include_extra_fields=True,
        _app_name=GELF_APP_NAME,
    )

    logger = logging.getLogger("gelf_producer")

    logger.setLevel(GELF_LOG_LEVEL)
    logger.addHandler(handler)

    print("Sending GELF messages...")

    count = 0
    alive = int(10 * 1000 / GELF_SLEEP_PERIOD_MS)

    while RUNNING:
        if count > 0 and count % alive == 0:
            print(f"Still sending logs, count: {count}...")

        extra = {
            "count": count,
            "rand": random.randint(0, 100),
            "nested": {"foo": {"bar": "baz"}},
        }

        logger.debug("This is a debug message.", extra=extra)
        logger.error("This is an error message.", extra=extra, exc_info=ValueError())
        logger.warning("This is a warning message.", extra=extra)
        logger.info("This is an info message.", extra=extra)

        count += 1
        time.sleep(GELF_SLEEP_PERIOD_MS / 1000)

    print("Finished sending messages.")


if __name__ == "__main__":
    main()
