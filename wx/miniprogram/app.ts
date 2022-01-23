import camelcaseKeys from 'camelcase-keys'
import { IAppOption } from './app-option'
import { auth } from './services/proto-gen/auth/auth-pb'

// app.ts
App<IAppOption>({
  globalData: {},
  onLaunch() {
    // 展示本地存储能力
    const logs = wx.getStorageSync('logs') || []
    logs.unshift(Date.now())
    wx.setStorageSync('logs', logs)

    // 登录
    wx.login({
      success: (res) => {
        wx.request({
          method: 'POST',
          url: 'http://localhost:8082/v1/auth/login',
          data: {
            code: res.code,
          } as auth.v1.ILoginRequest,
          dataType: 'json',
          responseType: 'text',
          success: (res) => {
            console.log(res.data)
            console.log('-----------------------')
            const loginRes = auth.v1.LoginResponse.fromObject(
              camelcaseKeys(res.data as object)
            )
            console.log(loginRes)
          },
        })
        console.log(res.code)
        // 发送 res.code 到后台换取 openId, sessionKey, unionId
      },
    })
  },
})
