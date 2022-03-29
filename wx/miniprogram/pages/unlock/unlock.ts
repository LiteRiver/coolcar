import { IAppOption } from '../../app-option'
import { TripService } from '../../services/trip'
import { consts } from '../../utils/consts'
import { routing } from '../../utils/routing'
import { wxapi } from '../../utils/wxapi'

const app = getApp<IAppOption>()

Page({
  carId: '',
  data: {
    shareLocation: false,
    avatarUrl: '',
  },
  onGetUserProfile() {
    wxapi.getUserProfile().then((res) => {
      app.resolveUserInfo(res.userInfo)
      this.setData({
        shareLocation: true,
      })
      wx.setStorageSync(consts.ShareLocationKey, true)
    })
  },
  async onLoad(opts: Record<'carId', string>) {
    const unlockOpts: routing.UnlockOpts = opts
    this.carId = unlockOpts.carId
    const userInfo = await app.globalData.userInfo
    this.setData({
      avatarUrl: userInfo.avatarUrl,
      shareLocation: wx.getStorageSync(consts.ShareLocationKey) || false,
    })
  },
  onShareLocationChanged(e: any) {
    const shareLocation: boolean = e.detail.value
    wx.setStorageSync(consts.ShareLocationKey, shareLocation)
    this.setData({
      shareLocation,
    })
  },
  onUnlockClicked() {
    wx.getLocation({
      type: 'gcj02',
      success: async (loc) => {
        console.log('starting a trip', {
          location: {
            latitude: loc.latitude,
            longitude: loc.longitude,
          },
          avatarUrl: this.data.shareLocation ? this.data.avatarUrl : '',
        })
        if (!this.carId) {
          console.error('no carId specified')
          return
        }

        wx.showLoading({
          title: '开锁中',
          mask: true,
        })

        try {
          const trip = await TripService.create({
            start: loc,
            carId: this.carId,
          })
          const tripId = trip.id
          wx.redirectTo({
            url: routing.driving({ tripId }),
            complete: () => {
              wx.hideLoading()
            },
          })
        } catch (err) {
          wx.hideLoading()
          wx.showToast({
            title: '创建行程失败',
            icon: 'none',
          })

          return
        }
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
