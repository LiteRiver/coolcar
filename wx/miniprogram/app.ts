import { IAppOption } from "./app-option"
import { getSetting, getUserInfo } from "./utils/util"

let resolveUserInfo: (
  value: WechatMiniprogram.UserInfo | PromiseLike<WechatMiniprogram.UserInfo>
) => void
let rejectUserInfo: (reason: any) => void

// app.ts
App<IAppOption>({
  globalData: {
    userInfo: new Promise((resolve, reject) => {
      resolveUserInfo = resolve
      rejectUserInfo = reject
    }),
  },
  async onLaunch() {
    // 展示本地存储能力
    const logs = wx.getStorageSync("logs") || []
    logs.unshift(Date.now())
    wx.setStorageSync("logs", logs)

    // getSetting()
    //   .then((res) => {
    //     if (res.authSetting["scope.userInfo"]) {
    //       return getUserInfo()
    //     }
    //     return undefined
    //   })
    //   .then((res) => {
    //     if (!res) {
    //       return
    //     }
    //     resolveUserInfo(res.userInfo)
    //   })
    //   .catch(rejectUserInfo)
  },
  resolveUserInfo(userInfo: WechatMiniprogram.UserInfo) {
    resolveUserInfo(userInfo)
  },
})
