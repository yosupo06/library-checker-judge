echo 'Install dlang'
wget https://github.com/ldc-developers/ldc/releases/download/v1.15.0/ldc2-1.15.0-linux-x86_64.tar.xz
tar -xf ldc2-1.15.0-linux-x86_64.tar.xz -C /opt
ln -s /opt/ldc2-1.15.0-linux-x86_64/bin/ldc2 /usr/bin/ldc2

echo 'Install Python(PyPy)'
wget https://bitbucket.org/pypy/pypy/downloads/pypy3.6-v7.1.1-linux64.tar.bz2
tar -xf pypy3.6-v7.1.1-linux64.tar.bz2 -C /opt
ln -s /opt/pypy3.6-v7.1.1-linux64/bin/pypy3 /usr/bin/pypy3