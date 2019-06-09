#!/usr/bin/env python3

import os
import sys
import shutil
import json
import glob
from subprocess import run, check_call, TimeoutExpired, CalledProcessError
from datetime import datetime
from logging import basicConfig, getLogger

basicConfig(
    filename='executer.log',
    level=os.getenv('LOG_LEVEL', 'DEBUG'),
    format="%(asctime)s %(levelname)s %(name)s :%(message)s"
)
logger = getLogger(__name__)

curdir = os.path.abspath(os.path.curdir)
sanddir = os.path.join(curdir, 'sand')
workdir = os.path.join(curdir, 'work')

logger.info('Launch executer.py')

logger.info('os env = {}'.format(os.environ))
run(['./prepare.sh'])


class Result:
    result = ''
    time = 0
    memory = 0

    def __init__(self, result, time, memory):
        self.result = result
        self.time = time
        self.memory = memory


def run_in_sandbox(execcmd, copyfiles, stdinpath, timelimit):
    result = {
        'status': 'IE',
        'time': -1,
        'memory': -1,
    }

    logger.info('execcmd: {}'.format(execcmd))
    check_call(['./prepare_exec.sh'])

    fstdin = None
    if stdinpath:
        logger.info('stdin: {}'.format(stdinpath))
        fstdin = open(stdinpath, 'r')

    for f in glob.glob('sand/*'):
        if os.path.isfile(f):
            os.remove(f)

    for f in copyfiles:
        shutil.copy(os.path.join(workdir, f), os.path.join(sanddir, f))
        
    try:
        cmd = ['cgexec', '-g', 'cpuset,memory:lib-judge',
               'chroot', '--userspec=library-checker-user:library-checker-user', 'sand']
        cmd.extend(execcmd.split())
        start = datetime.now()
        check_call(cmd, stdin=fstdin,
                   stdout=open('work/out.txt', 'w'), timeout=timelimit)
    except TimeoutExpired:
        result['status'] = 'TLE'
    except CalledProcessError:
        result['status'] = 'RE'
    else:
        end = datetime.now()
        result['status'] = 'OK'
        result['time'] = (end - start).seconds * 1000 + \
            (end - start).microseconds // 1000
        with open('/sys/fs/cgroup/memory/lib-judge/memory.max_usage_in_bytes', 'r') as f:
            result['memory'] = int(f.read())

    return result


while True:
    s = sys.stdin.readline().strip()
    logger.info('input: {}'.format(s))
    if s == 'last':
        break
    comm = json.load(open('work/comm.json', 'r'))
    logger.info('Command: {}'.format(comm))
    result = run_in_sandbox(comm['exec'],
                            copyfiles=comm.get('files', []),
                            stdinpath=comm.get('stdin', None),
                            timelimit=comm.get('timelimit', 2.0))
    logger.info('Result: {}'.format(result))
    with open('work/resp.json', 'w') as f:
        f.write(json.dumps(result))
    print('OK', flush=True)
