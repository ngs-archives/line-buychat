買い物チャット
==============


![QR Code](http://qr-official.line.me/L/BsqbfocuYK.png)

[![友達追加](https://scdn.line-apps.com/n/line_add_friends/btn/ja.png)](https://line.me/R/ti/p/%40xhe9481d)

[![CircleCI](https://circleci.com/gh/ngs/line-buychat/tree/master.svg?style=svg&circle-token=b93d2f1b5b11b10f45990807de1768ff7cac60ac)](https://circleci.com/gh/ngs/line-buychat/tree/master)

Setup
-----

```
# .envrc

## Grab LINE Credentials from
## https://developers.line.me/ba/
export LINE_CHANNEL_SECRET=...
export LINE_CHANNEL_TOKEN=...

## Grab AWS Credentials from
## https://console.aws.amazon.com/iam/home#/security_credential
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...

## Product Advertising Configurations from
## https://affiliate.amazon.co.jp/gp/associates/network/your-account/manage-tracking-ids.html
export AWS_PRODUCT_REGION=JP
export AWS_ASSOCIATE_TAG=buychat-22
```

License
-------

Copyright &copy; 2016 [LittleApps Inc.](https://littleapps.jp)
