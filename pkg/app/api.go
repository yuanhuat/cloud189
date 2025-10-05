package app

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gowsp/cloud189/pkg/invoker"
	"github.com/gowsp/cloud189/pkg/util"
)

type api struct {
	invoker *invoker.Invoker
	conf    *invoker.Config
}

func New(path string) *api {
	conf, _ := invoker.OpenConfig(path)
	api := &api{conf: conf}
	api.invoker = invoker.NewInvoker("https://api.cloud.189.cn", api.refresh, conf)
	api.invoker.SetPrepare(api.sign)
	return api
}

func Mem(username, password string) *api {
	conf := &invoker.Config{User: &invoker.User{Name: username, Password: password}}
	api := &api{conf: conf}
	api.invoker = invoker.NewInvoker("https://api.cloud.189.cn", api.refresh, conf)
	api.invoker.SetPrepare(api.sign)
	return api
}

func (api *api) refresh() error {
	s := api.conf.Session
	if s.Login() {
		params := url.Values{}
		params.Set("appId", "9317140619")
		params.Set("accessToken", s.AccessToken)
		var newSession invoker.Session
		if err := api.invoker.Post("/getSessionForPC.action", params, &newSession); err != nil {
			return err
		}
		s.Merge(newSession)
		return api.conf.Save()
	}
	user := api.conf.User
	if user.Name == "" || user.Password == "" {
		return errors.New("用户未登录或扫码不支持自动重新登录")
	}
	return api.PwdLogin(api.conf.User.Name, api.conf.User.Password)
}

func (api *api) sign(req *http.Request) {
	now := time.Now()
	query := req.URL.Query()
	// 填充客户端参数
	query.Set("rand", strconv.FormatInt(now.UnixMilli(), 10))
	query.Set("clientType", "TELEPC")
	query.Set("version", "7.1.8.0")
	query.Set("channelId", "web_cloud.189.cn")
	req.URL.RawQuery = query.Encode()

	// sha1(SessionKey=相应的值&Operate=相应值&RequestURI=相应值&Date=相应的值", SessionSecret)
	session := api.conf.Session
	if session.Empty() {
		return
	}
	date := now.Format(time.RFC1123)
	data := fmt.Sprintf("SessionKey=%s&Operate=%s&RequestURI=%s&Date=%s",
		session.Key, req.Method, req.URL.Path, date)
	// 追加上传参数
	if req.Host == "upload.cloud.189.cn" {
		data += "&params=" + query.Get("params")
	}
	req.Header.Set("Date", date)
	req.Header.Set("user-agent", "desktop")
	req.Header.Set("SessionKey", session.Key)
	req.Header.Set("Signature", util.Sha1(data, session.Secret))
	req.Header.Set("X-Request-ID", util.Random("xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx"))
}

// Logout 退出登录，清除session和配置
func (api *api) Logout() error {
	// 清除session信息
	if api.conf.Session != nil {
		api.conf.Session = &invoker.Session{}
	}

	// 清除用户信息
	if api.conf.User != nil {
		api.conf.User = &invoker.User{}
	}

	// 清除其他认证信息
	api.conf.SSON = ""
	api.conf.Auth = ""

	// 保存清空的配置到文件
	return api.conf.Save()
}
