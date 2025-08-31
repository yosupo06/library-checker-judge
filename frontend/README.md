# Library Checker Frontend

## Requirements

- nodejs18

## How to Use

```sh
npm install # or npm ci

# generate API client code from protoc
PROTO_PATH=../library-checker-judge/api/proto npm run protoc

# access to the API server of local (you must launch api server in local)
npm run dev
# access to the API server of judge.yosupo.jp
npm run dev -- --mode production
```

## Contributing

なんでも歓迎

### pre-commit hooks

このリポジトリはコミット前に Prettier/ESLint を実行する pre-commit をサポートします。

1. pre-commit をインストール: `pip install pre-commit`
2. Git フックを有効化: `pre-commit install`

以後、`src/**/*.ts(x)|js(x)` に変更があるコミットでは以下を実行し、失敗時はコミットが中断されます。
- `npm run prettier:check`
- `npm run lint`

## library-checker-project

- problems: [library-checker-problems](https://github.com/yosupo06/library-checker-problems)
- judge: [library-checker-judge](https://github.com/yosupo06/library-checker-judge)
- frontend: [library-checker-frontend](https://github.com/yosupo06/library-checker-frontend)
