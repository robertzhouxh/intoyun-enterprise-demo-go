# IntoYun 第三方服务器接入服务器开源实现(go版)
## 设计架构

```


                                     +-----+                                                        +-----+
                                     |cli-n|                                                        |dev-n|
                                     +-----+                                                        +-----+
                            +-----+     ^                                                  +-----+    |
                            |cli-2|     |                                                  |dev-2|    |
                            +--+--+     |                                                  +--+--+    |
                               ^        |                                                     |       |
                 +-----+       |        |                                       +-----+       |       |
                 |cli-0|       |        |                                       |dev-0|       |       |
                 +--+--+       |        |                                       +--+--+       |       |
                    ^          |        |                                          |          |       |
                    |          |        |                                          |          |       |
                    V          v        V                                          V          V       V
                  +----+-----+----+--------+                                   +-------------------------+
                  |    |     |    |   |    |                                   |                         |
                  |ch-0|     |ch-2|   |ch-n|                                   |                         |
                  |    |     |    |   |    |                                   |                         |
                  |----+-----+----+--------|                                   |                         |
                  |  ^          ^       ^  |                                   |                         |
                  |  |          |       |  |                                   |                         |
                  |========================|                                   |                         |
                  |                        |        RealTime Data              |                         |
                  |    Push Engine Layer   |<----------------------------------|                         |
                  |========================|                                   |                         |
                  |                        |        RestFul Api                |                         |
                  |                itysrv  |<--------------------------------->|             IntoYunCloud|
                  +------------------------+                                   +-------------------------+



```

1. 设备 dev-x 直接接入 IntoYun Cloud
2. 用户客户端 cli-x 通过 app 或 web 前端网页登陆第三方企业实现的后台服务器 itysrv,
   websocket 连接用来接收设备端的实时数据， http连接用来实现自己的业务逻辑以及通
   过 IntoYun Cloud 提供的 企业api 接口， 获取 IntoYun 平台的数据以及控制设备。
3. 设备端的实时数据会通过 IntoYun Cloud 直接转发到 itysrv 服务器中的 PushEngine，
   PushEnging 会根据 itysrv 实现的推送策略， 将这些实时数据推送到对应的 cli-x。
4. 整个 itysrv 由 http 服务 + websocket 服务 两部分组成， http 服务完成 web功能，
   websocket服务完成实时数据推送。
5. http 服务没有依赖开源web框架， 直接基于go 自带的库来实现。
6. itysrv 的 push engine layyer 完成了数据的aes 解密， 然后封装成自定义协议
   Proto（参考代码)推送到客户端cli-x, 客户端cli-x 解析之后即可获得设备端上报的实时原始数据。
7. 用户可以很快扩展自定义协议（libs/proto/proto.go)
8. 用户可以快速重写业务逻辑部分实现自定义业务逻辑(operator.go)。

## 跨平台

### Windows 安装 ###

Windows系统用户请按Win+R运行cmd，输入`systeminfo`后回车，稍等片刻，会出现一些系统信息。
在“系统类型”一行中，若显示“x64-based PC”，即为64位系统；若显示“X86-based PC”，则为32位系统。

访问 https://golang.org/dl/，32 位请选择名称中包含 windows-386 的 msi 安装包，64 位请选择名称中包含 windows-amd64 的。
下载好后运行，不要修改默认安装目录 C:\Go\，若安装到其他位置会导致不能执行自己所编写的 Go 代码。
安装完成后默认会在环境变量 Path 后添加 Go 安装目录下的 bin 目录 `C:\Go\bin\`，并添加环境变量 GOROOT，值为 Go 安装根目录 `C:\Go\` 。

**验证是否安装成功**

在运行中输入 `cmd` 打开命令行工具，在提示符下输入 `go`，检查是否能看到 Usage 信息。输入 `cd %GOROOT%`，看是否能进入 Go 安装目录。若都成功，说明安装成功。

不能的话请检查上述环境变量 Path 和 GOROOT 的值。若不存在请卸载后重新安装，存在请重启计算机后重试以上步骤。

## 修改配置文件

1. http 监听端口

    [base]
    http.addrs tcp@0.0.0.0:8081

2. websocket 监听端口

    [websocket]
    bind 0.0.0.0:8082

3. 服务器授权信息

    - appid  your-app-id
    - appsecret you-app-secret

4. kafka配置参数

    - topic device-data-${appid}
    - group ${appid}
    - kafka.list dps.intoyun.com:9092
    - sasl.enable true
    - sasl.user ${appid}
    - sasl.password ${appsecret}

5. log 参数配置
    - windows:  修改配置文件 srv.conf, 使用配置文件 itysrv-log-win.xml
    - linux 修改配置文件 srv.conf, 使用配置文件 itysrv-log.xml
    - macos 修改配置文件 srv.conf, 使用配置文件 itysrv-log.xml


## 运行服务(默认http:8081, websocket: 8082)

1. cd itysrv
2. go get -v
3. go build
4. ./itysrv

## 服务测试(默认连接websocket端口 8082, 创建1个客户端连接， 可以直接修改main.go 修改默认值)

1. cd itysrv/client
2. go build
3. ./client


测试程序默认会每10毫秒创建一个 websocket 连接, 创建两个, 可以修改main.go 来设置个数。
客户端通过认证之后， 每隔 30s 发一次心跳包, 目前服务器对客户端websocket连接的认
证没只是简单的直接按照 key-{0, 1}, 顺序创建订阅的key, 这样设备的实时数据转发到 itysrv 之后就可以根据
推送策略（目前全部推送到key-0 对应的cli-0) 完成实时数据的推送。

```
step1. 获取 token
    curl -X POST http://localhost:8081/v1/token

    6a08820093e15327cfe5ee13cd31d74f


step2. 获取企业所有的product
    curl --header "X-IntoYun-SrvToken:6a08820093e15327cfe5ee13cd31d74f" http://localhost:8081/v1/product

step3. 获取企业的某个product信息
    curl --header "X-IntoYun-SrvToken:6a08820093e15327cfe5ee13cd31d74f" http://localhost:8081/v1/product/${productId}

step4. 获取企业某个product产品下面的所有设备
    curl --header "X-IntoYun-SrvToken:6a08820093e15327cfe5ee13cd31d74f" http://localhost:8081/v1/device?productId=${productId}&page=1&size=20

step5. 控制某个设备
    curl -X POST --header "X-IntoYun-SrvToken:6a08820093e15327cfe5ee13cd31d74f" --header "Content-Type: application/json" -d '[{"dpId": 1, "type": "enum", "value": 0}]' http://localhost:8081/v1/control?productId=${productId}&deviceId=b

step5. 获取某个设备的历史数据
    curl --header "X-IntoYun-SrvToken:6a08820093e15327cfe5ee13cd31d74f" http://localhost:8081/v1/sensordata?productId=${productId}&start=1232555&end=2351233&deviceId=${devceid}&dpId=0&interval=10s
```


## 请注意 

在 client 中， 对于数据点的解析， 尤其针对 number 数据点类型的解析.
