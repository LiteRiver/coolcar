import { IAppOption } from "./app-option"
// import { getSetting, getUserInfo } from './utils/wxapi'

let resolveUserInfo: (
  value: WechatMiniprogram.UserInfo | PromiseLike<WechatMiniprogram.UserInfo>
) => void
let rejectUserInfo: (reason: any) => void

App<IAppOption>({
  globalData: {
    userInfo: new Promise((resolve, reject) => {
      resolveUserInfo = resolve
      rejectUserInfo = reject
    }),
  },
  onLaunch() {},
  resolveUserInfo(userInfo: WechatMiniprogram.UserInfo) {
    resolveUserInfo(userInfo)
  },
  rejectUserInfo(reason?: any) {
    rejectUserInfo(reason)
  },
})
