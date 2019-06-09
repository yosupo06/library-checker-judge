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


import argparse
import glob
import json
import os
from os.path import join
import shutil
import sys
import tempfile
import traceback
import zipfile
from datetime import datetime
from logging import basicConfig, getLogger
from subprocess import (DEVNULL, PIPE, CalledProcessError, Popen,
                        TimeoutExpired, check_call, run)
from time import sleep

import psycopg2
from termcolor import colored, cprint

basicConfig(
    level=os.getenv('LOG_LEVEL', 'DEBUG'),
    format="%(asctime)s %(levelname)s %(name)s :%(message)s"
)
logger = getLogger(__name__)

logger.info('Launch judge.py')

logger.info('Make workdir & sanddir')
curdir = os.path.abspath(os.path.curdir)
sanddir = join(curdir, 'sand')
workdir = join(curdir, 'work')
if os.path.exists(sanddir):
    shutil.rmtree(sanddir)
if os.path.exists(workdir):
    shutil.rmtree(workdir)
os.mkdir(sanddir)
os.mkdir(workdir)
os.chmod(sanddir, 0o777)
shutil.copy('testlib.h', 'work/testlib.h')

env = os.environ.copy()
env['POSTGRE_PASS'] = 'secret(off course, we should fix this code...)'
executer = Popen(['unshare', '-fpnm', '--mount-proc',
                  './executer.py'], stdin=PIPE, stdout=PIPE, env=env)


def run_in_sandbox(execcmd, copyfiles=[], stdinpath='', stdoutpath='', timelimit=2.0):
    data = {
        'exec': execcmd,
        'timelimit': timelimit,
    }
    if len(copyfiles):
        data['files'] = copyfiles
    if stdinpath:
        data['stdin'] = stdinpath
    if stdoutpath:
        data['stdout'] = stdoutpath

    logger.info('judge -> executer data: {}'.format(data))
    with open('work/comm.json', 'w') as f:
        f.write(json.dumps(data))
    executer.stdin.write(b'comm\n')
    executer.stdin.flush()

    s = executer.stdout.readline().decode('utf-8').strip()
    if s != 'OK':
        logger.error('Error executer: {}'.format(s))
        return {}
    logger.info('Return OK')
    return json.load(open('work/resp.json', 'r'))


def judge_single_case(execcmd, inpath, outpath, timelimit=2.0):
    new_inpath = os.path.join(workdir, 'case.in')
    new_outpath = os.path.join(workdir, 'case.out')
    shutil.copy(inpath, new_inpath)
    shutil.copy(outpath, new_outpath)
    inpath = new_inpath
    outpath = new_outpath
    anspath = os.path.join(workdir, 'case.ans')

    # run
    result = run_in_sandbox(execcmd, copyfiles=['main'],
                            stdinpath=inpath, stdoutpath=anspath, timelimit=timelimit)

    if result['status'] == 'OK':
        try:
            # output check
            run_in_sandbox('./checker case.in case.out case.ans',
                           copyfiles=['checker', 'case.in', 'case.out', 'case.ans'], timelimit=30.0)
        except CalledProcessError:
            result['status'] = 'WA'
        except TimeoutError:
            result['status'] = 'ITLE'
        else:
            result['status'] = 'AC'

    def get_color():
        status = result['status']
        if status == 'AC':
            return 'on_green'
        elif status == 'WA':
            return 'on_yellow'
        else:
            return 'on_red'

    logger.info('judged {} res={} {} msecs'.format(
        inpath, colored(result['status'], on_color=get_color()), result['time']))
    return result


def compile_checker():
    run_in_sandbox('g++ -O2 -std=c++14 -o checker checker.cpp',
                   copyfiles=['checker.cpp', 'testlib.h'], timelimit=30.0)
    shutil.copy(os.path.join(sanddir, 'checker'),
                os.path.join(workdir, 'checker'))


def compile(lang):
    assert(lang == 'cpp')  # todo
    run_in_sandbox('g++ -O2 -std=c++14 -o main main.cpp',
                   copyfiles=['main.cpp'], timelimit=30.0)
    shutil.copy(os.path.join(sanddir, 'main'), os.path.join(workdir, 'main'))


def fetchcases(conn, problemid):
    # get zip file
    testhash = ''
    with conn.cursor() as cursor:
        if cursor.execute('select testhash from problems where name = %s', (problemid, )) == 0:
            return
        testhash = cursor.fetchone()[0]

    zippath = os.path.join(workdir, 'cases-{}.zip'.format(testhash))

    if not os.path.exists(zippath):
        logger.info('Nothing {}, fetching...'.format(zippath))
        with conn.cursor() as cursor:
            if cursor.execute('select testzip from problems where name = %s', (problemid, )) == 0:
                return
            zipdata = cursor.fetchone()[0]
            with open(zippath, 'wb') as f:
                f.write(zipdata)

    logger.info('Extract zip file')

    indir = join(workdir, 'in')
    outdir = join(workdir, 'out')

    if os.path.exists('work/checker.cpp'):
        os.remove('work/checker.cpp')
    if os.path.exists(indir):
        shutil.rmtree(indir)

    if os.path.exists(outdir):
        shutil.rmtree(outdir)

    with zipfile.ZipFile(zippath, 'r') as f:
        f.extractall(workdir)


def judge(conn, subid):
    logger.info('Judge start submission id = {}'.format(subid))

    logger.info('Fetch data from SQL')
    submission = None
    with conn.cursor() as cursor:
        if cursor.execute('select problem, lang, source from submissions where id = %s', (subid, )) == 0:
            return
        submission = cursor.fetchone()

    fetchcases(conn, submission[0])

    # write source
    with open(os.path.join(workdir, 'main.cpp'), 'w') as f:
        f.write(submission[2])

    logger.info('Compile checker & source')
    compile_checker()
    compile('cpp')

    status = 'AC'
    time = -1
    memory = -1
    file_list = list(sorted(glob.glob('work/in/*')))
    for i, inpath in enumerate(file_list):
        _, filepath = os.path.split(inpath)
        name, _ = os.path.splitext(filepath)
        outpath = os.path.join('work/out', name + '.out')
        result = judge_single_case('./main', inpath=inpath, outpath=outpath)

        with conn.cursor() as cursor:
            cursor.execute('update submissions set status = %s where id = %s',
                           ('{}/{}'.format(i, len(file_list)), subid))
            conn.commit()

        if result['status'] != 'AC':
            status = result['status']
        time = max(time, result['time'])
        memory = max(memory, result['memory'])

    with conn.cursor() as cursor:
        cursor.execute('update submissions set status = %s, maxtime = %s, maxmemory = %s where id = %s',
                       (status, time, memory, subid))
        conn.commit()

    logger.info('End judge')


if __name__ == "__main__":
    logger.info('Connect SQL')
    hostname = os.environ.get('POSTGRE_HOST', '127.0.0.1')
    port = int(os.environ.get('POSTGRE_PORT', '5432'))
    user = os.environ.get('POSTGRE_USER', 'postgres')
    password = os.environ.get('POSTGRE_PASS', 'passwd')
    conn = psycopg2.connect(
        host=hostname,
        port=port,
        user=user,
        password=password,
        database='librarychecker'
    )
    while True:
        sleep(1)
        sql = 'select id from queue'
        res = None
        with conn.cursor() as cursor:
            cursor.execute('select id, submission from tasks limit 1')
            res = cursor.fetchone()
            if res == None:
                continue
            cursor.execute('delete from tasks where id = %s', (res[0],))
            conn.commit()

        subid = res[1]
        try:
            judge(conn, subid)
        except Exception as e:
            ex, ms, tb = sys.exc_info()
            logger.error("Unexpected error: {}".format(traceback.print_tb(tb)))
            with conn.cursor() as cursor:
                cursor.execute('update submissions set status = %s where id = %s',
                               ('IE', subid))
                conn.commit()
