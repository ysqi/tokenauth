# tokenauth
   tokenauth 是一个 token 维护验证存储包，支持单点和多点模式，默认使用 boltdbf 存储 token 数据。

# 安装

`go get github.com/ysqi/tokenauth`

也可以通过`-u`参数来更新tokenauth 和所依赖的第三方包

`go get -u github.com/ysqi/tokenauth`

# 功能

+ 支持自定义存储
+ 默认使用boltdbf存储token到本地
+ 随机生成客户令牌
+ 自定义算法生成 Token
+ 支持对一个客户维护N个Token
+ 支持对一个客户仅维护一个 Token
+ 支持Token有效性验证

# 使用场景

__为第三方客户端颁发Token__

作为使用平台，第三方可注册使用平台服务，当客户端登录成功后单一客户端拉取一个全新Token，后续客户端可直接携带 Token访问平台资源，而不需要提供账号和密码信息，同时可在Token到期后请求拉取新 Token。

__单点登录__

用户在第一次成功登录后，可给用户拉取一个全新Token，同时该用户相关的旧Token立即失效。用户可以使用该 Token 访问其他子系统，如 App 登录后，可 URL 携带 Token 访问 Web 站点资源。

# 简单使用

```Go
import (
	"fmt"
	"github.com/ysqi/tokenauth"
)
func main() {

	if err := tokenauth.UseDeaultStore(); err != nil {
		panic(err)
	}
	defer tokenauth.Store.Close()

	// Ready.
	d := &tokenauth.DefaultProvider{}
	globalClient := tokenauth.NewAudienceNotStore("globalClient", d.GenerateSecretString)

	//New token
	token, err := tokenauth.NewSingleToken("singleID", globalClient, d.GenerateTokenString)
	if err != nil {
		fmt.Println("generate token fail,", err.Error())
		return
	}
	 Check token
	if checkToken, err := tokenauth.ValidateToken(token.Value); err != nil {
		fmt.Println("token check did not pass,", err.Error())
	} else {
		fmt.Println("token check pass,token Expiration date:", checkToken.DeadLine)
	}

}
```

1.程序初始化时，需手动选择Store方案

+ 选择默认方案:
```go
tokenauth.UseDeaultStore();
```
+ 选择自定义Store
```go
if store, err := tokenauth.NewStore(newStoreName, storeConf); err != nil {
	panic(err)
}else if err = tokenauth.ChangeTokenStore(store); err != nil {
	panic(err)
}
```

2.定义生成密钥 Secret 和 Token 算法

+ 选择默认算法:
```go
d := &tokenauth.DefaultProvider{}
secretFunc := d.GenerateSecretString
tokenFunc := d.GenerateTokenString
```

+ 选择自定义算法
```go
secretFunc := func(clientID string) (secretString string) { return "myself secret for all client" }
tokenFunc := func(audience *Audience) string { return "same token string" }
```

3.使用算法在 Store 中创建存储一个听众（相当于用户）
```go
client := tokenauth.NewAudienceNotStore("client name", secretFunc)
```

4.使用算法给用户颁发一个或多个Token
```go
token, err := tokenauth.NewToken(client, tokenFunc)
if err != nil {
	fmt.Println("generate token fail,", err.Error())
}

// more ...
t2 ,err  := tokenauth.NewToken(client, tokenFunc)
```
> 每个Token 都要自己的生命周期,`Store`自动定期清除过期`Token`，默认有效时常为:`tokenauth.TokenPeriod`

5.验证 Token String 的有效性
```go
if checkToken, err := tokenauth.ValidateToken(tokenString); err != nil {
	fmt.Println("token check did not pass,", err.Error())
} else {
	fmt.Println("token check pass,token Expiration date:", checkToken.DeadLine)
}
```
6.当然可以主动删除 Token
```go
err:=tokenauth.Store.DeleteToken(tokenString)
```
