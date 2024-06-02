package utility

import (
	"strings"

	"github.com/cool-team-official/cool-admin-go/cool"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

type CarInfo struct {
	Carid         string
	Email         string
	IsPlus        bool
	IsPlusStr     string
	RefreshCookie string
	AccessToken   string
	Password      string
	PlanType      string
}

func CheckCar(ctx g.Ctx, carid string) (carInfo *CarInfo, err error) {
	// g.Log().Info(ctx, "check carid:", carid)
	sessionVar, err := cool.CacheManager.Get(ctx, "session:"+carid)
	if err != nil {
		return
	}
	sessionJson := gjson.New(sessionVar)
	// sessionJson.Dump()
	carInfo = &CarInfo{}
	carInfo.Carid = carid
	email := sessionJson.Get("user.email").String()
	if email == "" {
		err = gerror.New("email is empty")
		return
	}
	carInfo.Email = email
	// password := sessionJson.Get("user.password").String()
	// if password == "" {
	// 	err = gerror.New("password is empty")
	// 	return
	// }
	// carInfo.Password = password

	refreshCookie := sessionJson.Get("refreshCookie").String()
	if refreshCookie == "" {
		err = gerror.New("refreshCookie is empty")
		return
	}
	carInfo.RefreshCookie = refreshCookie
	accessToken := sessionJson.Get("accessToken").String()
	if accessToken == "" {
		err = gerror.New("accessToken is empty")
		return
	}
	carInfo.AccessToken = accessToken
	// models := sessionJson.Get("models").Array()
	// isPlus := len(models) > 1
	plan_type := sessionJson.Get("accountCheckInfo.plan_type").String()
	if plan_type != "" {
		carInfo.PlanType = plan_type
	}
	carInfo.IsPlus = strings.Contains(sessionJson.String(), "32767")
	if carInfo.IsPlus {
		carInfo.IsPlusStr = "PLUS"
	} else {
		carInfo.IsPlusStr = "4o"
	}
	return
}

// CloseCar is a function that closes a car.
func CloseCar(ctx g.Ctx, carid string) (err error) {
	_, err = cool.CacheManager.Remove(ctx, "session:"+carid)
	return
}
