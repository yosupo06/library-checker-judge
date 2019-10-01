echo 'Install dlang'
wget https://github.com/ldc-developers/ldc/releases/download/v1.15.0/ldc2-1.15.0-linux-x86_64.tar.xz
tar -xf ldc2-1.15.0-linux-x86_64.tar.xz

echo 'Install Rust'
wget https://static.rust-lang.org/dist/rust-1.35.0-x86_64-unknown-linux-gnu.tar.gz
tar -xf rust-1.35.0-x86_64-unknown-linux-gnu.tar.gz
mkdir rust
./rust-1.35.0-x86_64-unknown-linux-gnu/install.sh --prefix=./rust

echo 'Install Python(PyPy)'
wget https://bitbucket.org/pypy/pypy/downloads/pypy3.6-v7.1.1-linux64.tar.bz2
tar -xf pypy3.6-v7.1.1-linux64.tar.bz2
