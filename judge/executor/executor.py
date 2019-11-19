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


import json
import sys
import resource
import argparse

from os import getenv, getuid, environ
from pwd import getpwnam
from shutil import rmtree
from pathlib import Path
from logging import basicConfig, getLogger
import subprocess
from subprocess import run
import tempfile
from psutil import Process
from time import perf_counter

logger = getLogger(__name__)


def outside(args, cmd):
    logger.info('outside')
    tmp = tempfile.NamedTemporaryFile()

    core = Path(sys.argv[0]).parent / 'executor_core'

    if not core.exists():
        logger.warn('compile executor_core.cpp: start')
        core_src = Path(sys.argv[0]).parent / 'executor_core.cpp'
        subprocess.check_call(['g++', str(core_src), '-o', str(core)])
        logger.warn('compile executor_core.cpp: finished')

    arg = ['unshare', '-fpnm', '--mount-proc']
    arg += [sys.argv[0]]
    arg += ['--inside']
    arg += ['--insideresult', tmp.name]
    arg += sys.argv[1:]

    mycode = 0
    returncode = -1
    time = -1
    memory = -1

    try:
        proc = subprocess.run(arg, timeout=args.tl + 5.0)
        mycode = proc.returncode
        result = json.load(tmp)
        returncode = result.get('returncode', -1)
        time = result.get('time', -1)
        memory = result.get('memory', -1)
    except subprocess.TimeoutExpired:
        logger.warning('outside catch timeout, this is unexpected')
        mycode = 124
        time = args.tl

    result = {
        'returncode': returncode,
        'time': time,
        'memory': memory,
    }
    logger.info('outside result = {}'.format(result))
    if args.result:
        args.result.write(json.dumps(result))
    return mycode


def inside(args, execcmd):
    logger.info('inside execute: {}'.format(execcmd))

    # expand stack
    resource.setrlimit(resource.RLIMIT_STACK,
                       (resource.RLIM_INFINITY, resource.RLIM_INFINITY))
    # TODO: use TemporaryDirectory
    tmpdir = Path(tempfile.mkdtemp())
    tmpdir.chmod(0o777)
    prepare_mount(tmpdir, args.overlay)
    prepare_cgroup()

    core = Path(sys.argv[0]).parent / 'executor_core'
    cmd = [core, 'time.txt', 'cgexec', '-g', 'pids,cpuset,memory:lib-judge']
    cmd += ['chroot',
            '--userspec=library-checker-user:library-checker-user', str(tmpdir)]
    cmd += ['sh', '-c', ' '.join(['cd', 'sand', '&&'] + execcmd)]

    mycode = 0
    returncode = -1
    time = -1
    memory = -1

    env = environ.copy()
    env["HOME"] = "/home/library-checker-user"

    try:
        proc = subprocess.run(cmd,
                              stdin=args.stdin,
                              stdout=args.stdout,
                              stderr=args.stderr,
                              env=env,
                              timeout=args.tl)
        returncode = proc.returncode
    except subprocess.TimeoutExpired:
        logger.info('timeout command')
        mycode = 124  # error code of timeout command
        time = args.tl
    else:
        with open('time.txt', 'r') as f:
            time = float(f.read())
        with open('/sys/fs/cgroup/memory/lib-judge/memory.max_usage_in_bytes', 'r') as f:
            memory = int(f.read())

    subprocess.run(['pkill', '-KILL', '-u', 'library-checker-user'])
    for child in Process().children():
        child.wait()

    result = {
        'returncode': returncode,
        'time': time,
        'memory': memory,
    }
    logger.info('inside result = {}'.format(result))

    args.insideresult.write(json.dumps({
        'returncode': returncode,
        'time': time,
        'memory': memory,
    }))
    return mycode


def prepare_mount(tmpdir: Path, overlay):
    sanddir = tmpdir / 'sand'
    sanddir.mkdir()
    if overlay:
        workdir = Path(tempfile.mkdtemp())
        workdir.chmod(0o777)
        upperdir = Path(tempfile.mkdtemp())
        upperdir.chmod(0o777)
        cmd = ['mount', '-t', 'overlay', 'overlay', '-o']
        cmd += ['lowerdir={},upperdir={},workdir={}'.format(
            './', str(upperdir), str(workdir))]
        cmd += [str(sanddir)]
        subprocess.run(cmd, check=True)
    else:
        cmd = ['mount', '--bind', './', str(sanddir)]
        subprocess.run(cmd, check=True)

    (tmpdir / 'proc').mkdir()
    subprocess.run(['mount', '-t', 'proc', 'none',
                    str(tmpdir / 'proc')], check=True)

    (tmpdir / 'tmp').mkdir()
    (tmpdir / 'tmp').chmod(0o777)
    for dname in ['dev', 'sys', 'bin', 'lib', 'lib64', 'usr', 'etc', 'opt', 'var', 'home']:
        (tmpdir / dname).mkdir()
        subprocess.run(['mount', '--bind', '-o', 'ro', '/' + dname,
                        str(tmpdir / dname)], check=True)


def prepare_cgroup():
    run(['cgdelete', 'pids,cpuset,memory:/lib-judge'])
    run(['cgcreate', '-g', 'pids,cpuset,memory:/lib-judge'], check=True)
    run(['cgset', '-r', 'pids.max=1000', 'lib-judge'], check=True)
    run(['cgset', '-r', 'cpuset.cpus=0', 'lib-judge'], check=True)
    run(['cgset', '-r', 'cpuset.mems=0', 'lib-judge'], check=True)
    run(['cgset', '-r', 'memory.limit_in_bytes=1G', 'lib-judge'], check=True)
    run(['cgset', '-r', 'memory.memsw.limit_in_bytes=1G', 'lib-judge'], check=True)


if __name__ == "__main__":
    basicConfig(
        level=getenv('LOG_LEVEL', 'WARN'),
        format="%(asctime)s %(levelname)s %(name)s : %(message)s"
    )
    assert sys.argv.count("--") == 1
    if sys.argv.count("--") == 0:
        logger.error('args must have -- : {}'.format(sys.argv))
        exit(1)
    sep_index = sys.argv.index("--")

    parser = argparse.ArgumentParser(
        description='Testcase Generator', usage='%(prog)s [options] -- command')
    parser.add_argument('--stdin', type=argparse.FileType('r'), help='stdin')
    parser.add_argument('--stdout', type=argparse.FileType('w'), help='stdout')
    parser.add_argument('--stderr', type=argparse.FileType('w'), help='stderr')
    parser.add_argument('--overlay', action='store_true',
                        help='overlay current dir?')
    parser.add_argument('--result', type=argparse.FileType('w'),
                        help='result file')
    parser.add_argument('--inside', action='store_true',
                        help='inside flag(DONT USE THIS FLAG DIRECTLY)')
    parser.add_argument('--insideresult', type=argparse.FileType('w'),
                        help='inside result file(DONT USE THIS FLAG DIRECTLY)')
    parser.add_argument('--tl', type=float, help='Time Limit', default=3600.0)

    args = parser.parse_args(sys.argv[1:sep_index])
    cmd = sys.argv[sep_index + 1:]

    if not (0 <= args.tl and args.tl <= 3600):
        logger.error('invalid tl: {}'.format(args.tl))
        exit(1)

    if args.inside:
        exit(inside(args, cmd))
    else:
        exit(outside(args, cmd))
