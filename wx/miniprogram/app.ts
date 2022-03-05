import { IAppOption } from './app-option'
import { Coolcar } from './services/request'

// app.ts
App<IAppOption>({
  globalData: {},
  onLaunch() {
    // 展示本地存储能力
    const logs = wx.getStorageSync('logs') || []
    logs.unshift(Date.now())
    wx.setStorageSync('logs', logs)

    // 登录
    // Coolcar.login()
  },
})
