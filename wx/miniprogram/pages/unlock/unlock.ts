import { IAppOption } from '../../app-option'
import { getUserProfile } from '../../utils/wxapi'

const app = getApp<IAppOption>()
const ShareLocationKey = 'share-location'

Page({
  data: {
    shareLocation: false,
    avatarUrl: '',
  },
  onGetUserProfile() {
    getUserProfile().then((res) => {
      app.resolveUserInfo(res.userInfo)
      this.setData({
        shareLocation: true,
      })
      wx.setStorageSync(ShareLocationKey, true)
    })
  },
  async onLoad() {
    const userInfo = await app.globalData.userInfo
    this.setData({
      avatarUrl: userInfo.avatarUrl,
      shareLocation: wx.getStorageSync(ShareLocationKey) || false,
    })
  },
  onShareLocationChanged(e: any) {
    const shareLocation: boolean = e.detail.value
    wx.setStorageSync(ShareLocationKey, shareLocation)
    this.setData({
      shareLocation,
    })
  },
  onUnlockClicked() {
    wx.getLocation({
      type: 'gcj02',
      success: (loc) => {
        console.log('starting a trip', {
          location: {
            latitude: loc.latitude,
            longitude: loc.longitude,
          },
          avatarUrl: this.data.shareLocation ? this.data.avatarUrl : '',
        })
        wx.showLoading({
          title: '开锁中',
          mask: true,
        })
        setTimeout(() => {
          wx.redirectTo({
            url: '/pages/driving/driving',
            complete: () => {
              wx.hideLoading()
            },
          })
        }, 2000)
      },
      fail: () => {
        wx.showToast({
          icon: 'none',
          title: '请前往设置页授权位置信息',
        })
      },
    })
  },
})
