# Library Checker Frontend

## Requirements

- npm
- docker

## How to Use

```sh
./gen_protoc.sh /path/to/library-checker-judge/api/proto/library_checker.proto

npm install

# access to the API server of judge.yosupo.jp
npm run start

# access to the API server of local (you must launch api server in local)
REACT_APP_API_URL=http://localhost:58080 npm run start
```

## Contributing

なんでも歓迎

## library-checker-project

- problems: [library-checker-problems](https://github.com/yosupo06/library-checker-problems)
- judge: [library-checker-judge](https://github.com/yosupo06/library-checker-judge)
- frontend: [library-checker-frontend](https://github.com/yosupo06/library-checker-frontend)
