#!/usr/bin/env python3

import os, toml, glob
import tempfile, zipfile
import hashlib
import subprocess
import psycopg2 #pip3 install psycopg2


print('[*] deploy problem to SQL')

print('[*] make case')
# make case
tomlpath = os.path.abspath('../library-checker-problems/problems.toml')
tomldir = os.path.dirname(tomlpath)
subprocess.check_call(['../library-checker-problems/generate.py', tomlpath])
problems = toml.load(tomlpath)

print('[*] connect SQL')
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


for problem in problems['Problems']:
    print('[*] deploy {}'.format(problem['Name']))
    probdir = os.path.join(tomldir, problem['Dir'])
    with tempfile.NamedTemporaryFile(suffix='.zip') as tmp:
        with zipfile.ZipFile(tmp.name, 'w') as newzip:
            for f in sorted(glob.glob(probdir + '/in/*.in')):
                print(f, ' ', os.path.relpath(f, probdir))
                newzip.write(f, arcname=os.path.relpath(f, probdir))
            for f in sorted(glob.glob(probdir + '/out/*.out')):
                print(f, ' ', os.path.relpath(f, probdir))
                newzip.write(f, arcname=os.path.relpath(f, probdir))

        tmp.seek(0)

        name = problem['Name']
        data = tmp.read()
        m = hashlib.sha256()
        m.update(data)
        datahash = m.hexdigest()

        print(len(data), datahash)

        sql = 'insert into problems (name, testhash, testzip) values (%s, %s, %s)'
        with conn.cursor() as cursor:
            cursor.execute(sql, (name, datahash, data))
        conn.commit()
conn.close()

# upload problems
