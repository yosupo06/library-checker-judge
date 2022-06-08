#!/usr/bin/env python3

import argparse
from pathlib import Path
from logging import Logger, getLogger, basicConfig
from subprocess import run
from typing import List
from os import getenv

logger: Logger = getLogger(__name__)


def create_instance(name: str, zone: str, env: str, preemptible: bool):
    args = ['gcloud', 'compute', 'instances', 'create']
    args += [name]
    args += ['--zone', zone]
    args += ['--machine-type', 'c2-standard-4']
    args += ['--boot-disk-size', '50GB']
    args += ['--boot-disk-type', 'pd-ssd']
    args += ['--image-family', '{}-judge-image-family'.format(env)]
    args += ['--service-account', 'gce-judge@library-checker-project.iam.gserviceaccount.com']
    args += ['--scopes', 'default,cloud-platform']
    if preemptible:
        args += ['--preemptible']

    run(args, check=True)

def run_in_instance(name: str, zone: str, args: List[str]):
    args2 = ['gcloud', 'compute', 'ssh', 'root@{}'.format(name)]
    args2 += ['--zone', zone]
    args2 += ['--']
    args2 += args
    run(args2, check=True)

def send_file(src: Path, name: str, zone: str, dst: Path):
    args = ['gcloud', 'compute', 'scp']
    args += [str(src.absolute())]
    args += ['root@{}:{}'.format(name, str(dst))]
    args += ['--zone', zone]
    run(args, check=True)


if __name__ == '__main__':
    basicConfig(
        level=getenv('LOG_LEVEL', 'INFO'),
    )
    parser = argparse.ArgumentParser()
    parser.add_argument('--name', required=True)
    parser.add_argument('--zone', required=True)
    parser.add_argument('--env', required=True)
    parser.add_argument('--preemptible', action='store_true')

    args = parser.parse_args()

    name: str = args.name
    zone: str = args.zone
    env: str = args.env
    preemptible: bool = args.preemptible

    logger.info('create instance, name = %s, zone = %s, env = %s, preemptible = %d', name, zone, env, preemptible)
    create_instance(name, zone, env, preemptible)
