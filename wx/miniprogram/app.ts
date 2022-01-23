import camelcaseKeys from 'camelcase-keys'
import { IAppOption } from './app-option'
import { auth } from './services/proto-gen/auth/auth-pb'
import { rental } from './services/proto-gen/rental/rental-pb'

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
          url: 'http://localhost:8081/v1/auth/login',
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

            wx.request({
              url: 'http://localhost:8081/v1/rental/trip',
              method: 'POST',
              data: {
                start: 'abc',
              } as rental.v1.CreateTripRequest,
              header: {
                authorization: `Bearer ${loginRes.accessToken}`,
              },
            })
          },
        })
        console.log(res.code)
        // 发送 res.code 到后台换取 openId, sessionKey, unionId
      },
    })
  },
})
