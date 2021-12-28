# packer

## Build Judge Image for GCP

```
gcloud auth application-default login
cd packer
packer build -var 'env=test' . # for testing
packer build -var 'env=prod' . # for production
```

Library Checkerで稼働するジャッジサーバーのイメージは[packer](https://www.packer.io/)でビルドされている。
