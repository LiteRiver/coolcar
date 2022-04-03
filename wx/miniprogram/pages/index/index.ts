import { IAppOption } from '../../app-option'
import { ProfileService } from '../../services/profile'
import { rental } from '../../services/proto-gen/rental/rental-pb'
import { TripService } from '../../services/trip'
import { routing } from '../../utils/routing'

Page({
  isPageShowing: false,
  avatarUrl: '',
  carLocation: {
    latitude: 23.099994,
    longitude: 113.32452,
  },
  data: {
    location: {
      latitude: 31,
      longitude: 120,
    },
    settings: {
      showCompass: false,
      showLocation: true,
      scale: 10,
      enableZoom: true,
      enableScroll: true,
      enable3D: false,
      enableOverlooking: false,
    },
    markers: [
      {
        iconPath: '/images/car.png',
        id: 0,
        latitude: 23.099994,
        longitude: 113.32452,
        width: 50,
        height: 50,
      },
      {
        iconPath: '/images/car.png',
        id: 0,
        latitude: 23.099994,
        longitude: 114.32452,
        width: 50,
        height: 50,
      },
    ],
  },
  async onLoad() {
    const userInfo = await getApp<IAppOption>().globalData.userInfo
    this.setData({
      avatarUrl: userInfo.avatarUrl,
    })
  },
  onMyLocationTap() {
    wx.getLocation({
      type: 'gcj02',
      success: (res) => {
        this.setData({
          location: {
            latitude: res.latitude,
            longitude: res.longitude,
          },
        })
      },
      fail: () => {
        wx.showToast({
          icon: 'none',
          title: '请前往设置页授权',
        })
      },
    })
  },
  onMyTripsClicked() {
    wx.navigateTo({
      url: routing.mytrips(),
    })
  },
  moveCars() {
    const map = wx.createMapContext('map')

    const moveCar = () => {
      this.carLocation.latitude += 0.1
      this.carLocation.longitude += 0.1
      map.translateMarker({
        destination: {
          latitude: this.carLocation.latitude,
          longitude: this.carLocation.longitude,
        },
        markerId: 0,
        duration: 5000,
        rotate: 0,
        autoRotate: false,
        animationEnd: () => {
          if (this.isPageShowing) {
            moveCar()
          }
        },
      })
    }

    moveCar()
  },
  async onScanClicked() {
    const res = await TripService.list(rental.v1.TripStatus.IN_PROGRESS)
    if (res.trips[0]) {
      await this.selectComponent('#tripModal').showModal()
      wx.navigateTo({
        url: routing.driving({
          tripId: res.trips[0].id!,
        }),
      })
      return
    }
    wx.scanCode({
      success: async () => {
        const carId = 'car123'
        const unlockUrl = routing.unlock({ carId })

        const profile = await ProfileService.get()
        if (profile.identityStatus === rental.v1.IdentityStatus.VERIFIED) {
          wx.navigateTo({
            url: unlockUrl,
          })
        } else {
          await this.selectComponent('#licModal').showModal()
          wx.navigateTo({
            url: routing.registration({ redirectURL: unlockUrl }),
          })
        }
      },
      fail: console.error,
    })
  },
})
