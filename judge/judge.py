#!/usr/bin/env python3

import argparse
import glob
import json
import os
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
sanddir = os.path.join(curdir, 'sand')
workdir = os.path.join(curdir, 'work')
if os.path.exists(sanddir):
    shutil.rmtree(sanddir)
if os.path.exists(workdir):
    shutil.rmtree(workdir)
os.mkdir(sanddir)
os.mkdir(workdir)
os.chmod(sanddir, 0o777)


env = os.environ.copy()
env['POSTGRE_PASS'] = 'secret(of course, we should fix this code...)'
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


def judgecase(execcmd, inpath, outpath, timelimit=2.0):
    anspath = os.path.join(workdir, 'ans.txt')

    # run
    result = run_in_sandbox(
        execcmd, copyfiles=['main'], stdinpath=inpath, timelimit=timelimit)

    shutil.copy(os.path.join(workdir, 'out.txt'), anspath)

    color = ''
    if result['status'] == 'OK':
        try:
            # output check
            check_call(['diff', anspath, outpath], stdout=DEVNULL)
        except CalledProcessError:
            result['status'] = 'WA'
            color = 'on_yellow'
        else:
            result['status'] = 'AC'
            color = 'on_green'
    else:
        color = 'on_red'

    logger.info('judged {} res={} {} msecs'.format(
        inpath, colored(result['status'], on_color=color), result['time']))
    return result


def compilecxx(srcpath):
    run_in_sandbox('g++ -O2 -std=c++14 -o main main.cpp',
                   copyfiles=['main.cpp'], timelimit=20.0)
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

    return zippath


def judge(conn, subid):
    logger.info('Judge start submission id = {}'.format(subid))

    logger.info('Fetch data from SQL')
    submission = None
    problem = None
    with conn.cursor() as cursor:
        if cursor.execute('select problem, lang, source from submissions where id = %s', (subid, )) == 0:
            return
        submission = cursor.fetchone()

    # write source
    with open(os.path.join(workdir, 'main.cpp'), 'w') as f:
        f.write(submission[2])

    zippath = fetchcases(conn, submission[0])

    srcpath = os.path.join(workdir, 'main.cpp')  # Todo: other lang

    logger.info('Extract zip file')

    indir = os.path.join(workdir, 'in')
    outdir = os.path.join(workdir, 'out')

    if os.path.exists(indir):
        shutil.rmtree(indir)

    if os.path.exists(outdir):
        shutil.rmtree(outdir)

    with zipfile.ZipFile(zippath, 'r') as f:
        f.extractall(workdir)

    logger.info('Compile main.cpp')
    compilecxx(srcpath)

    status = 'AC'
    consume_time = -1
    consume_memory = -1
    file_list = list(sorted(glob.glob(indir + '/*')))
    for i, inpath in enumerate(file_list):
        _, filepath = os.path.split(inpath)
        name, _ = os.path.splitext(filepath)
        result = judgecase(
            './main', inpath, os.path.join(outdir, name + '.out'))

        with conn.cursor() as cursor:
            cursor.execute('update submissions set status = %s where id = %s',
                           ('{}/{}'.format(i, len(file_list)), subid))
            conn.commit()

        if result['status'] != 'AC':
            status = result['status']
        consume_time = max(consume_time, result['time'])
        consume_memory = max(consume_memory, result['memory'])

    with conn.cursor() as cursor:
        cursor.execute('update submissions set status = %s, maxtime = %s, maxmemory = %s where id = %s',
                       (status, consume_time, consume_memory, subid))
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
