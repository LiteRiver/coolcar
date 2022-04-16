import { IAppOption } from '../../app-option'
import { CarService } from '../../services/car'
import { car } from '../../services/proto-gen/car/car-pb'
import { TripService } from '../../services/trip'
import { consts } from '../../utils/consts'
import { routing } from '../../utils/routing'
import { wxapi } from '../../utils/wxapi'

const app = getApp<IAppOption>()

Page({
  carRefresher: 0,
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
    this.data.shareLocation = e.detail.value
    wx.setStorageSync(consts.ShareLocationKey, this.data.shareLocation)
  },
  onUnlockClicked() {
    wx.getLocation({
      type: 'gcj02',
      success: async (loc) => {
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
            avatarUrl: this.data.shareLocation ? this.data.avatarUrl : '',
          })
          const tripId = trip.id
          this.carRefresher = setInterval(async () => {
            const c = await CarService.get(this.carId)
            if (c.status === car.v1.CarStatus.UNLOCKED) {
              this.clearCarRefresher()
              wx.redirectTo({
                url: routing.driving({ tripId }),
                complete: () => {
                  wx.hideLoading()
                },
              })
            }
          }, 2000)
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
  onUnload() {
    this.clearCarRefresher()
    wx.hideLoading()
  },
  clearCarRefresher() {
    if (this.carRefresher) {
      clearInterval(this.carRefresher)
      this.carRefresher = 0
    }
  },
})
