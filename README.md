# library-checker-project

- problems: [library-checker-problems](https://github.com/yosupo06/library-checker-problems)
- judge: [library-checker-judge](https://github.com/yosupo06/library-checker-judge)
- frontend: [library-checker-frontend](https://github.com/yosupo06/library-checker-frontend)

# Recommended Environment

- Ubuntu 18.04
- ~~Debian 9(stretch)~~

# 動かすのに必要なもの

- docker

```
sudo apt install postgresql-client libpq-dev python3 python3-dev python3-pip g++ cgroup-tools libcap2-bin

pip3 install termcolor toml psycopg2 psutil
```

など

# 準備


### cgroupでmemory swapを管理する
/etc/default/grubに以下を書き、reboot
```
GRUB_CMDLINE_LINUX="swapaccount=1"
```

- References: https://unix.stackexchange.com/questions/147158/how-to-enable-swap-accounting-for-memory-cgroup-in-archlinux


### ジャッジ用のシステムユーザーを作成する

```
sudo useradd library-checker-user -u 990 -r -s /sbin/nologin -M
```

postgreコンテナはデフォルトで999をUIDとして使う、一方でuseradd -rはデフォルトでUID 999のユーザーを作る。
そしてジャッジはpkill -u library-checker-user(このユーザーのプロセスを全部消す)を行うため、UIDが被ってはいけない。

どちらかを変更すること

# Local Test

- library-checker-problems / library-chcker-judgeは同じディレクトリにcloneする
- library-checker-frontendはどこでもよい, go getとかするとよい？(TODO)

### SQL立ち上げ
```
cd /your/path/of/library-checker-judge
./launch_local
```

dockerでpostgre SQLが立ち上がり、問題データが生成され、SQLに格納される

dockerコマンドにsudoが必要な場合、`./launch_local`をsudoで実行する必要がある。
しかしlibrary-checker-problems内にいろいろrootでフォルダが作られて面倒な事になるので、非推奨

dockerグループに自分を登録することでsudoなしでdockerが使えるようになる
- References: https://qiita.com/DQNEO/items/da5df074c48b012152ee

### Launch Judge
```
cd /your/path/of/library-checker-judge/judge
sudo ./judge.py
```

### Launch web server

```
cd /your/path/of/library-checker-problems/
go run main.go
```
