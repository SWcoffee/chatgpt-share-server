package auth

import (
	"backend/config"
	"backend/modules/chatgpt/model"
	"backend/utility"
	"strings"
	"sync"
	"time"

	"github.com/cool-team-official/cool-admin-go/cool"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gmlock"
	"github.com/gogf/gf/v2/text/gstr"
)

var (
	errSessionStr = `
	{
		"detail": {
			"type":    "invalid_request_error",
			"param":   <nil>,
			"code":    "session_invalidated",
			"message": "Your authentication token has expired. Please try signing in again.",
		},
	}`
	carLocks sync.Map
)

func Session(r *ghttp.Request) {
	ctx := r.GetCtx()
	usertoken := r.Session.MustGet("usertoken").String()
	var sessionJson *gjson.Json

	carid := r.Session.MustGet("carid").String()
	carinfo, err := utility.CheckCar(ctx, carid)
	if err != nil {
		g.Log().Error(ctx, err)
		r.Response.WriteJson(g.Map{
			"code": 0,
			"msg":  "服务器错误",
		})
		return
	}
	if gmlock.TryLock(carid + "-refreshSession") {
		defer gmlock.Unlock(carid + "-refreshSession")

		getsessionUrl := config.LOGINPROXY + "/refreshsession"
		getsessionVar := g.Client().SetHeader("authkey", config.AUTHKEY).PostVar(ctx, getsessionUrl, g.MapStrStr{
			"email":         carinfo.Email,
			"refreshCookie": carinfo.RefreshCookie,
			"authkey":       config.AUTHKEY,
		})
		sessionJson = gjson.New(getsessionVar)
		sessionJson.Dump()

		detail := sessionJson.Get("detail").String()
		if gstr.Contains(detail, "RefreshAccessTokenError") {
			utility.CloseCar(ctx, carid)
			r.Response.Status = 401
			r.Response.WriteJson(gjson.New(errSessionStr))
			return
		}
		detail = sessionJson.Get("detail").String()
		if gstr.Contains(detail, "RefreshAccessTokenError") || gstr.Contains(detail, "token_invalidated") {
			cool.DBM(model.NewChatgptSession()).Where("email=?", carinfo.Email).Update(g.Map{
				"status":          0,
				"officialSession": "账号刷新失败,请重新登录",
			})
			utility.CloseCar(ctx, carid)
			r.Response.Status = 401
			r.Response.WriteJson(gjson.New(errSessionStr))
			return
		}
		email := sessionJson.Get("user.email").String()
		if email == "" {
			// 先使用缓存中的session
			sessionVar, err := cool.CacheManager.Get(ctx, "session:"+carid)
			if err != nil {
				r.Response.Status = 401
				r.Response.WriteJson(gjson.New(errSessionStr))
				return
			}
			sessionJson = gjson.New(sessionVar)
			// 移除sessionJson中的refreshCookie
			sessionJson.Remove("refreshCookie")
			// 移除sessionJson中的models
			sessionJson.Remove("models")
			sessionJson.Set("user.email", "share@openai.com")
			sessionJson.Set("user.name", carid)
			sessionJson.Set("user.image", "/avatars.png")
			sessionJson.Set("user.picture", "/avatars.png")
			sessionJson.Set("user.id", "user-"+usertoken)
			sessionJson.Set("accessToken", usertoken)

			r.Response.WriteJson(sessionJson)
			return
		}
		//models := sessionJson.Get("models").Array()
		// isPlus := len(models) > 1
		isPlus := strings.Contains(sessionJson.String(), "32767")
		plan_type := sessionJson.Get("accountCheckInfo.plan_type").String()
		if plan_type == "free" {
			isPlus = false
		}
		// 更新账号信息
		cool.DBM(model.NewChatgptSession()).Where("email=?", email).Update(g.Map{
			"officialSession": sessionJson.String(),
			"isPlus":          isPlus,
			"status":          1,
		})
		// 更新缓存
		cool.CacheManager.Set(ctx, "session:"+carid, sessionJson.String(), 90*24*time.Hour)
	} else {
		// 先使用缓存中的session
		sessionVar, err := cool.CacheManager.Get(ctx, "session:"+carid)
		if err != nil {
			r.Response.Status = 401
			r.Response.WriteJson(gjson.New(errSessionStr))
			return
		}
		sessionJson = gjson.New(sessionVar)

	}
	// 移除sessionJson中的refreshCookie
	sessionJson.Remove("refreshCookie")
	// 移除sessionJson中的models
	sessionJson.Remove("models")
	sessionJson.Set("user.email", "share@openai.com")
	sessionJson.Set("user.name", carid)
	sessionJson.Set("user.image", "/avatars.png")
	sessionJson.Set("user.picture", "/avatars.png")
	sessionJson.Set("user.id", "user-"+usertoken)
	sessionJson.Set("accessToken", usertoken)

	r.Response.WriteJson(sessionJson)
}
