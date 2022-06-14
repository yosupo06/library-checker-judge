#!/usr/bin/env python3

import argparse
import random
import string
from logging import Logger, getLogger, basicConfig
from subprocess import CalledProcessError, run
from pathlib import Path
from os import getenv
from time import sleep

from create_instance import create_instance, delete_instance, run_in_instance, send_file

logger: Logger = getLogger(__name__)

def build_judge():
    args = ['go', 'test', '-c']
    run(args, check=True, cwd='../judge')

def run_test(name: str, zone: str, env: str, remain: bool):
    build_judge()
    send_file(Path('../judge/judge.test'), name, zone, Path('/root/judge.test'))
    send_file(Path('../langs/langs.toml'), name,
              zone, Path('/root/langs.toml'))
    send_file(Path('../judge/sources/testlib.h'), name, zone, Path('/root/testlib.h'))

    args = ['/root/judge.test']
    args += ['-langs', '/root/langs.toml']
    run_in_instance(name, zone, args)


if __name__ == '__main__':
    basicConfig(
        level=getenv('LOG_LEVEL', 'INFO'),
    )

    parser = argparse.ArgumentParser()
    parser.add_argument('--zone', required=True)
    parser.add_argument('--env', required=True)
    parser.add_argument('--remain', action='store_true')

    args = parser.parse_args()

    name: str = 'library-checker-judge-test-' + ''.join(
        random.choices(string.ascii_lowercase, k=10))
    zone: str = args.zone
    env: str = args.env
    remain: bool = args.remain

    try:
        create_instance(name, zone, env, preemptible=True)
        run_test(name, zone, env, remain)
    except Exception as e:
        logger.error('error {}'.format(e))
        raise e
    finally:
        if not remain:
            delete_instance(name, zone)
