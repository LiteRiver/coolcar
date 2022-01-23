package wechat

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Remote struct {
	AppId  string
	Secret string
}

type OpenIdResponse struct {
	OpenId     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionId    string `json:"unionid"`
	ErrCode    int32  `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

var baseUrl string = "https://api.weixin.qq.com/sns/jscode2session"

func (svc *Remote) getOpenIdUrl(code string) string {
	params := url.Values{}
	params.Set("appid", svc.AppId)
	params.Set("secret", svc.Secret)
	params.Set("js_code", code)
	params.Set("grant_type", "authorization_code")

	return strings.Join([]string{baseUrl, params.Encode()}, "?")
}

func (svc *Remote) GetOpenId(code string) (*OpenIdResponse, error) {
	url := svc.getOpenIdUrl(code)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var openIdRes OpenIdResponse
	err = json.Unmarshal(body, &openIdRes)
	if err != nil {
		return nil, err
	}

	if openIdRes.ErrCode != 0 {
		return nil, fmt.Errorf("%s", openIdRes.ErrMsg)
	}

	return &openIdRes, nil
}
