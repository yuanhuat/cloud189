修改代码后需要重新编译文件
编译方法；go build -o cloud189.exe ./cmd/cloud189
运行方法；./cloud189.exe web
api文档在docs\html\
常用api调用方法在README.md

二维码登录的逻辑：`qr_login.go` 中二维码扫描检测的方法和逻辑如下：

## 核心检测机制
### 1. 轮询检测方式
程序使用 定时轮询 的方式检测二维码状态：

- 每隔 3秒 发送一次状态查询请求
- 使用 time.NewTicker(3 * time.Second) 创建定时器
### 2. 状态查询接口
`query` 方法负责查询二维码状态：

```
func (c *QrCodeReq) query(conf 
*appConf) qrCodeState {
    req, _ := http.NewRequest(http.
    MethodPost, "https://open.e.189.
    cn/api/logbox/oauth2/
    qrcodeLoginState.do", nil)
    // ... 设置请求参数
    resp, _ := http.DefaultClient.Do
    (req)
    var status qrCodeState
    json.NewDecoder(resp.Body).
    Decode(&status)
    return status
}
```
### 3. 状态码判断逻辑
在 `QrLogin` 方法的无限循环中，通过 status.Status 字段判断二维码状态：

```
switch status.Status {
case -106:
    log.Println("not 
    scanned")        // 未扫描
case -11002:
    log.Println
    ("unconfirmed")        // 已扫描
    但未确认
case 0:
    log.Println
    ("logged")             // 登录成
    功
    return &LoginResult{ToUrl: 
    status.RedirectUrl, SSON: 
    status.SSON}, nil
default:
    return nil, errors.New("unknown 
    status")  // 未知状态，返回错误
}
```
## 状态码含义
状态码 含义 说明 -106 未扫描 二维码还没有被手机扫描 -11002 未确认 二维码已被扫描，但用户还没有在手机上确认登录 0 登录成功 用户已确认登录，可以获取登录凭证 其他 未知状态 可能是错误或超时等异常情况

## 工作流程
1. 1.
   生成二维码 ：调用 getUUID.do 接口获取二维码UUID
2. 2.
   显示二维码URL ：输出二维码图片链接供用户扫描
3. 3.
   开始轮询 ：每3秒调用一次 qrcodeLoginState.do 接口
4. 4.
   状态判断 ：根据返回的状态码判断当前扫描状态
5. 5.
   循环等待 ：直到状态变为成功(0)或出现错误才退出循环