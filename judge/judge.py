#!/usr/bin/env python3

import os, argparse
import psycopg2
import shutil
import glob
from time import sleep
from subprocess import check_call, TimeoutExpired, CalledProcessError, DEVNULL
from termcolor import colored, cprint
from datetime import datetime

print('[*] launch judge.py')

curdir = os.path.abspath(os.path.curdir)

# connect sql
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

class Result:
    result = ''
    time = 0
    memory = 0
    def __init__(self, result, time, memory):
        self.result = result
        self.time = time
        self.memory = memory

# TODO: output_checker
def judgecase(execfile, inpath, anspath, timelimit = 2.0):
    outpath = './tmp.out'

    # run
    res = ''
    color = ''
    start = datetime.now()

    try:
        check_call([execfile], stdin=open(inpath, 'r'), stdout=open(outpath, 'w'), timeout=timelimit)
    except TimeoutExpired:
        res = 'TLE'
        color = 'on_blue'
    except CalledProcessError:
        res = 'RE'
        color = 'on_red'        
    else:
        end = datetime.now()
        try:
            # output check
            check_call(['diff', outpath, anspath], stdout=DEVNULL)
        except CalledProcessError:
            res = 'RE'
            color = 'on_yellow'
        else:
            res = 'AC'
            color = 'on_green'

    end = datetime.now()
    usemsec = (end - start).seconds * 1000 + (end - start).microseconds // 1000

    print('[*] judged {} res={} {} msecs'.format(inpath, colored(res, on_color=color), usemsec))
    return Result(res, usemsec, -1)

# source must be abspath
def compilecxx(srcpath):
	dirname, srcname = os.path.split(srcpath)
	srctitle, _ = os.path.splitext(srcname)
	os.chdir(dirname)
	check_call(['g++', '-O2', '-std=c++14',
		'-I', os.path.join(curdir, 'common'),
		srcname, '-o', srctitle])

def judge(subid):    
    print('[!] judge start id = {}'.format(subid))
    submittion = None
    problem = None
    with conn.cursor() as cursor:
        if cursor.execute('select problem, lang, source from submittions where id = %s', (subid, )) == 0:
            return        
        submittion = cursor.fetchone()

    with conn.cursor() as cursor:
        if cursor.execute('select testzip source from problems where name = %s', (submittion[0], )) == 0:
            return
        problem = cursor.fetchone()

    with open('main.cpp', 'w') as f:
        f.write(submittion[2])
    
    with open('cases.zip', 'wb') as f:
        f.write(problem[0])

    # unzip    
    if os.path.exists('in'):
        shutil.rmtree('in')

    if os.path.exists('out'):
        shutil.rmtree('out')
    check_call(['unzip', 'cases.zip'])


    print('[!] end fetch data')

    print('[*] compile main.cpp')
    compilecxx(os.path.abspath('main.cpp'))

    status = 'AC'
    consume_time = -1
    consume_memory = -1
    file_list = list(sorted(glob.glob(curdir + '/in/*')))
    for i, inpath in enumerate(file_list):
        _, filepath = os.path.split(inpath)
        name, _ = os.path.splitext(filepath)
        result = judgecase(os.path.abspath('main'), inpath, curdir + '/out/' + name + '.out')

        with conn.cursor() as cursor:
            cursor.execute('update submittions set status = %s where id = %s',
                ('{}/{}'.format(i, len(file_list)), subid))
            conn.commit()

        if result.result != 'AC':
            status = result.result
        consume_time = max(consume_time, result.time)
        consume_memory = max(consume_memory, result.memory)

    with conn.cursor() as cursor:
        cursor.execute('update submittions set status = %s, maxtime = %s, maxmemory = %s where id = %s',
            (status, consume_time, consume_memory, subid))
        conn.commit()

    print('[!] end judge')


while True:
    sleep(1)
    sql = 'select id from queue'
    res = None
    with conn.cursor() as cursor:
        cursor.execute('select id, submittion from tasks limit 1')
        res = cursor.fetchone()
        if res == None:
            print('waiting...')
            continue            
        cursor.execute('delete from tasks where id = %s', (res[0],))
        conn.commit()

    judge(res[1])


conn.close()
