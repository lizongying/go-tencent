# go tencent

腾讯云ssl证书自动申请、安装

* [go-tencent](https://github.com/lizongying/go-tencent)

## dev

```shell
export GOPROXY=https://mirrors.tencent.com/go/

go install github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common@latest
go install github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl@latest
```

### build

```shell
make ssl
```

## run

* -secret-id set tencent secret-id，default ""
* -secret-key set tencent secret-key，default ""
* -region set region，default ""
* -save-dir set certificates save path，default "/tmp/"

```shell
# Change to your own secret-id & secret-key and set env on unix-like
export TENCENT_SECRET_ID=secret-id
export TENCENT_SECRET_KEY=secret-key

./releases/tencent_ssl_linux_amd64 -save-dir /etc/nginx/
```

## todo

* 网络等错误兼容处理
* 默认只安装到nginx，其他的请自行修改
* 腾讯云的其他功能

### 赞赏

![image](./appreciate.jpeg)