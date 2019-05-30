#!/usr/bin/env python3

import os, argparse
import pymysql
import time

print('[*] launch judge.py')

# connect sql
hostname = os.environ.get('MYSQL_HOST', '127.0.0.1')
port = int(os.environ.get('MYSQL_PORT', '3306'))
user = os.environ.get('MYSQL_USER', 'root')
password = os.environ.get('MYSQL_PASS', 'passwd')

conn = pymysql.connect(
    host=hostname,
    port=port,
    user=user,
    password=password,
    database='librarychecker'
)

def judge(subid):    
    print('[!] judge start id = {}'.format(subid))
    submittion = None
    problem = None
    with conn.cursor() as cursor:
        if cursor.execute('select problem, lang, source from submittion where id = %s', (subid, )) == 0:
            return        
        submittion = cursor.fetchone()

    with conn.cursor() as cursor:
        if cursor.execute('select testzip source from problem where name = %s', (submittion[0], )) == 0:
            return
        problem = cursor.fetchone()

    with open('main.cpp', 'w') as f:
        f.write(submittion[2])
    
    with open('cases.zip', 'wb') as f:
        f.write(problem[0])
    
    print('[!] end judge')


while True:
    sql = 'select id from queue'
    res = None
    with conn.cursor() as cursor:
        conn.begin()
        if cursor.execute('select id, submittion from queue limit 1') == 0:
            conn.commit()
            time.sleep(1)
            continue            
        res = cursor.fetchone()
        cursor.execute('delete from queue where id = %s', (res[0],))
        conn.commit()

    judge(res[1])


conn.close()
