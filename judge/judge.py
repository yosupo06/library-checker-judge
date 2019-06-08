#!/usr/bin/env python3

import argparse
import glob
import os
import shutil
import json
import tempfile
import zipfile
from datetime import datetime
from logging import basicConfig, getLogger
from subprocess import (DEVNULL, CalledProcessError, Popen, TimeoutExpired,
                        check_call, run, PIPE)
from time import sleep

import psycopg2
from termcolor import colored, cprint

basicConfig(
    level = os.getenv('LOG_LEVEL', 'DEBUG'),
    format = "%(asctime)s %(levelname)s %(name)s :%(message)s"
)
logger = getLogger(__name__)

curdir = os.path.abspath(os.path.curdir)
sanddir = os.path.join(curdir, 'sand')
workdir = os.path.join(curdir, 'work')

logger.info('Launch judge.py')

#executer = Popen(['unshare', '-m', './executer.py'], stdin=PIPE, stdout=PIPE)
#executer = Popen(['unshare', '-pm', './executer.py'], stdin=PIPE, stdout=PIPE)
executer = Popen(['unshare', '-fpnm', './executer.py'], stdin=PIPE, stdout=PIPE)

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

def run_in_sandbox(execcmd, stdinpath='', stdoutpath='', timelimit=2.0):
    data = {
        'exec': execcmd,
        'timelimit': timelimit
    }
    if stdinpath:
        data['stdin'] = stdinpath
    if stdoutpath:
        data['stdout'] = stdoutpath

    logger.info('judge -> executer data: {}'.format(data))
    with open('comm.json', 'w') as f:
        f.write(json.dumps(data))
    executer.stdin.write(b'comm\n')
    executer.stdin.flush()

    s = executer.stdout.readline().decode('utf-8').strip()
    if s != 'OK':
        logger.error('Error executer: {}'.format(s))
        return {}
    logger.info('Return OK')
    return json.load(open('resp.json', 'r'))


def judgecase(execcmd, inpath, outpath, timelimit=2.0):
    anspath = os.path.join(workdir, 'ans.txt')

    # run
    result = run_in_sandbox(execcmd, stdinpath=inpath, timelimit=timelimit)

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
    shutil.copy(srcpath, os.path.join(sanddir, 'main.cpp'))
    run_in_sandbox('g++ -O2 -std=c++14 -o main main.cpp', timelimit=20.0)


def judge(subid):
    logger.info('Judge start submittion id = {}'.format(subid))

    logger.info('Fetch data from SQL')
    submittion = None
    problem = None
    with conn.cursor() as cursor:
        if cursor.execute('select problem, lang, source from submittions where id = %s', (subid, )) == 0:
            return
        submittion = cursor.fetchone()

    with conn.cursor() as cursor:
        if cursor.execute('select testzip from problems where name = %s', (submittion[0], )) == 0:
            return
        problem = cursor.fetchone()

    logger.info('Extact fetched data')
    srcpath = os.path.join(workdir, 'main.cpp')  # Todo: other lang
    zippath = os.path.join(workdir, 'cases.zip')
    indir = os.path.join(workdir, 'in')
    outdir = os.path.join(workdir, 'out')
    with open(os.path.join(workdir, 'main.cpp'), 'w') as f:
        f.write(submittion[2])

    with open(zippath, 'wb') as f:
        f.write(problem[0])

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
            cursor.execute('update submittions set status = %s where id = %s',
                           ('{}/{}'.format(i, len(file_list)), subid))
            conn.commit()

        if result['status'] != 'AC':
            status = result['status']
        consume_time = max(consume_time, result['time'])
        consume_memory = max(consume_memory, result['memory'])

    with conn.cursor() as cursor:
        cursor.execute('update submittions set status = %s, maxtime = %s, maxmemory = %s where id = %s',
                       (status, consume_time, consume_memory, subid))
        conn.commit()

    logger.info('End judge')


# read sql queue and judge
while True:
    sleep(1)
    sql = 'select id from queue'
    res = None
    with conn.cursor() as cursor:
        cursor.execute('select id, submittion from tasks limit 1')
        res = cursor.fetchone()
        if res == None:
            continue
        cursor.execute('delete from tasks where id = %s', (res[0],))
        conn.commit()

    judge(res[1])


conn.close()

executer.wait()
