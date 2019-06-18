#!/usr/bin/env python3
# Copyright 2019 Kohei Morita
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#     http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


import glob
import json
import sys

from os import getenv, getuid, environ
from pwd import getpwnam
from shutil import rmtree, copy
from pathlib import Path
from datetime import datetime
from logging import basicConfig, getLogger
from subprocess import CalledProcessError, TimeoutExpired, Popen, run, DEVNULL
from psutil import Process
basicConfig(
    filename='executer.log',
    level=getenv('LOG_LEVEL', 'DEBUG'),
    format="%(asctime)s %(levelname)s %(name)s :%(message)s"
)
logger = getLogger(__name__)

curdir: Path = Path.cwd()
sanddir: Path = Path.cwd() / 'sand'

with open('../compiler/PATH.txt') as f:
    path = f.readline().strip()
    environ['PATH'] = path + ':' + environ['PATH']

def run_in_sandbox(execcmd, copyfiles, stdin, stdout, timelimit):
    status = 'IE'
    time = -1
    memory = -1

    logger.info('execcmd: {}'.format(execcmd))

    for f in sanddir.glob('*'):        
        if f.is_file():
            f.unlink()
        elif f.stat().st_uid == getpwnam('library-checker-user')[2]:
            rmtree(f)

    for f in copyfiles:
        fp = Path(f)
        copy(fp, sanddir / fp.name)

    run(['./prepare_exec.sh'], check=True)
    cmd = ['cgexec', '-g', 'cpuset,memory:lib-judge',
           'chroot', '--userspec=library-checker-user:library-checker-user', 'sand']
    cmd.extend(execcmd.split())

    start = datetime.now()
    proc = Popen(cmd,
                 stdin=stdin,
                 stdout=stdout)
                 #stderr=DEVNULL)
    try:
        proc.wait(timeout=timelimit)
    except TimeoutExpired:
        status = 'TLE'
    except CalledProcessError:
        status = 'RE'
    else:
        end = datetime.now()
        if proc.returncode:
            status = 'RE'
        else:
            status = 'OK'
        time = (end - start).seconds * 1000 + \
            (end - start).microseconds // 1000
        with open('/sys/fs/cgroup/memory/lib-judge/memory.max_usage_in_bytes', 'r') as f:
            memory = int(f.read())
    
    run(['pkill', '-KILL', '-u', 'library-checker-user'])
    for child in Process().children():
        child.wait()
        
    return {
        'status': status,
        'time': time,
        'memory': memory,
    }


if __name__ == "__main__":
    logger.info('Launch executer.py')
    run(['./prepare.sh'])

    while True:
        s = sys.stdin.readline().strip()
        logger.info('input: {}'.format(s))
        if s == 'last':
            break

        comm = json.load(open('work/comm.json', 'r'))
        logger.info('Command: {}'.format(comm))

        stdinpath = comm.get('stdin', None)
        stdin = DEVNULL
        if stdinpath:
            stdin = open(stdinpath, 'r')
        stdoutpath = comm.get('stdout', None)        
        stdout = DEVNULL
        if stdoutpath:
            stdout = open(stdoutpath, 'w')

        result = run_in_sandbox(comm['exec'],
                                copyfiles=comm.get('files', []),
                                stdin=stdin,
                                stdout=stdout,
                                timelimit=comm.get('timelimit', 2.0))
        logger.info('Result: {}'.format(result))
        with open('work/resp.json', 'w') as f:
            f.write(json.dumps(result))
        print('OK', flush=True)
