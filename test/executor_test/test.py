#!/usr/bin/env python3

import unittest
from logging import basicConfig, getLogger
from os import getenv
from pathlib import Path
from shutil import copy
from subprocess import run, PIPE
from tempfile import TemporaryDirectory, NamedTemporaryFile
import json

executor = Path('../../judge/v2/executor.py').absolute()
logger = getLogger(__name__)


def get_tmpdir(src: Path):
    tmpdir = TemporaryDirectory()
    Path(tmpdir.name).chmod(0o777)
    copy(src, Path(tmpdir.name) / src)
    return tmpdir


def get_result(cmd, cwd, overlay):
    logger.info('execute {}'.format(cmd))
    with NamedTemporaryFile() as resfile:
        cmd = [executor, '--result', resfile.name] + cmd
        if overlay:
            cmd = cmd + ['--overlay']
        run(cmd, cwd=cwd, check=True)
        result = json.load(resfile)
        logger.info('result {}'.format(result))
        return result


class TestHelloWorld(unittest.TestCase):
    def test_cpp(self):
        tmpdir = get_tmpdir('Hello.cpp')
        result = get_result(['g++', 'Hello.cpp'], tmpdir.name, False)
        self.assertEqual(result['returncode'], 0)
        result = get_result(['./a.out'], tmpdir.name, True)
        self.assertEqual(result['returncode'], 0)
        self.assertLess(result['time'], 0.05)
        tmpdir.cleanup()

    def test_java(self):
        tmpdir = get_tmpdir('Hello.java')
        result = get_result(['javac', 'Hello.java'], tmpdir.name, False)
        self.assertEqual(result['returncode'], 0)
        result = get_result(['java', 'Hello'], tmpdir.name, True)
        self.assertEqual(result['returncode'], 0)
        tmpdir.cleanup()

    def test_java_twoclasses(self):
        tmpdir = get_tmpdir('TwoClasses.java')
        result = get_result(['javac', 'TwoClasses.java'], tmpdir.name, False)
        self.assertEqual(result['returncode'], 0)
        result = get_result(['java', 'TwoClasses'], tmpdir.name, True)
        self.assertEqual(result['returncode'], 0)
        tmpdir.cleanup()

    def test_python(self):
        tmpdir = get_tmpdir('Hello.py')
        result = get_result(['python3', 'Hello.py'], tmpdir.name, True)
        self.assertEqual(result['returncode'], 0)
        tmpdir.cleanup()


class TestTLE(unittest.TestCase):
    def test_tle(self):
        tmpdir = get_tmpdir('TLE.cpp')
        result = get_result(['g++', 'TLE.cpp'], tmpdir.name, False)
        self.assertEqual(result['returncode'], 0)
        result = get_result(['./a.out', '--tl', '2.0'], tmpdir.name, True)
        self.assertEqual(result['returncode'], 124)
        self.assertAlmostEqual(result['time'], 2.0)
        tmpdir.cleanup()

if __name__ == "__main__":
    basicConfig(
        level=getenv('LOG_LEVEL', 'DEBUG'),
        format="%(asctime)s %(levelname)s %(name)s : %(message)s"
    )
    unittest.main()
