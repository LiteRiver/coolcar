package main

import (
	"encoding/json"
	"fmt"
)

type OpenIdResponse struct {
	OpenId     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionId    string `json:"unionid"`
	ErrCode    int32  `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func main() {
	str := `{"session_key":"JFSocTxQMTxrqDxZ6PNbBA==","openid":"oYgrr5DYALhYUazmk6AnU_yxokq4", "errcode": 3, "errmsg": "err", "unionid": "unionid-test"}`
	res := OpenIdResponse{}
	fmt.Println(res.ErrCode)
	err := json.Unmarshal([]byte(str), &res)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%T\n", res.ErrCode)

	bytes, err := json.Marshal(&res)
	if err != nil {
		panic(err)
	}

	fmt.Println("------------------------------------")
	fmt.Printf("%s\n", bytes)
}
