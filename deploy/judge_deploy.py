#!/usr/bin/env python3

import argparse
import random
import string
from logging import Logger, getLogger, basicConfig
from subprocess import CalledProcessError, run
from pathlib import Path
from os import getenv
from time import sleep

from create_instance import create_instance, run_in_instance, send_file

logger: Logger = getLogger(__name__)

def build_judge():
    args = ['go', 'build', '../judge']
    run(args, check=True, cwd='../judge')


if __name__ == '__main__':
    basicConfig(
        level=getenv('LOG_LEVEL', 'INFO'),
    )

    parser = argparse.ArgumentParser()
    parser.add_argument('--zone', required=True)
    parser.add_argument('--env', required=True)
    parser.add_argument('--preemptible', action='store_true')

    args = parser.parse_args()

    name: str = 'lib-judge-executor-' + ''.join(
        random.choices(string.ascii_lowercase, k=10))
    zone: str = args.zone
    env: str = args.env
    preemptible: bool = args.preemptible

    logger.info('create instance, name = %s, zone = %s, env = %s, preemptible = %s',
                name, zone, env, preemptible)
    create_instance(name, zone, env, preemptible)

    while True:
        try:
            run_in_instance(name, zone, ['echo', 'connected'])
        except CalledProcessError:
            logger.info('failed to connect...')
        else:
            break
        sleep(10)

    build_judge()
    send_file(Path('../judge/judge'), name, zone, Path('/root/judge'))
    send_file(Path('../langs/langs.toml'), name, zone, Path('/root/langs.toml'))
    send_file(Path('../judge/testlib.h'), name, zone, Path('/root/testlib.h'))
    run_in_instance(name, zone, ['mkdir', '-p', '/usr/local/lib/systemd/system'])
    send_file(Path('./judge.service'), name,
              zone, Path('/usr/local/lib/systemd/system/judge.service'))
    run_in_instance(name, zone, ['systemctl', 'daemon-reload'])
    run_in_instance(name, zone, ['service', 'judge', 'start'])
