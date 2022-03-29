export namespace wxapi {
  export function getSetting(): Promise<WechatMiniprogram.GetSettingSuccessCallbackResult> {
    return new Promise((resolve, reject) => {
      wx.getSetting({
        success: resolve,
        fail: reject,
      })
    })
  }

  export function getUserProfile(): Promise<WechatMiniprogram.GetUserProfileSuccessCallbackResult> {
    return new Promise((resolve, reject) => {
      wx.getUserProfile({
        desc: '显示头像',
        success: resolve,
        fail: reject,
      })
    })
  }

  export function getUserInfo(): Promise<WechatMiniprogram.GetUserInfoSuccessCallbackResult> {
    return new Promise((resolve, reject) => {
      wx.getUserInfo({
        success: resolve,
        fail: reject,
      })
    })
  }
}
