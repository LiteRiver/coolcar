import { IAppOption } from '../../app-option'
import { CarService } from '../../services/car'
import { ProfileService } from '../../services/profile'
import { rental } from '../../services/proto-gen/rental/rental-pb'
import { TripService } from '../../services/trip'
import { routing } from '../../utils/routing'

interface Marker {
  iconPath: string
  id: number
  latitude: number
  longitude: number
  width: number
  height: number
}

const defaultAvatar = '/images/car.png'
const initialLat = 30
const initialLng = 120

Page({
  socket: undefined as WechatMiniprogram.SocketTask | undefined,
  isPageShowing: false,
  avatarUrl: '',
  data: {
    location: {
      latitude: initialLat,
      longitude: initialLng,
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
    markers: [] as Marker[],
  },
  async onLoad() {
    const userInfo = await getApp<IAppOption>().globalData.userInfo
    this.setData({
      avatarUrl: userInfo.avatarUrl,
    })
  },
  onShow() {
    this.isPageShowing = true
    if (!this.socket) {
      this.setData(
        {
          markers: [],
        },
        () => {
          this.setUpCarPosUpdater()
        }
      )
    }
  },

  onHide() {
    this.isPageShowing = false
    if (this.socket) {
      this.socket.close({
        success: () => {
          this.socket = undefined
        },
      })
    }
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
  setUpCarPosUpdater() {
    const map = wx.createMapContext('map')
    const markersByCarId = new Map<string, Marker>()
    let translating = false
    const endTranslation = () => {
      translating = false
    }
    this.socket = CarService.subscribe((car) => {
      if (!car.id || translating || !this.isPageShowing) {
        console.log('dropped')
        return
      }
      const marker = markersByCarId.get(car.id)
      if (!marker) {
        const newMarker = {
          id: this.data.markers.length,
          iconPath: car.car?.driver?.avatarUrl || defaultAvatar,
          latitude: car.car?.position?.latitude || initialLat,
          longitude: car.car?.position?.longitude || initialLng,
          width: 20,
          height: 20,
        }
        markersByCarId.set(car.id, newMarker)
        this.data.markers.push(newMarker)
        translating = true
        this.setData(
          {
            markers: this.data.markers,
          },
          endTranslation
        )
        return
      }

      const newLat = car.car?.position?.latitude || initialLat
      const newLng = car.car?.position?.longitude || initialLng

      const newAvatar = car.car?.driver?.avatarUrl || defaultAvatar
      if (marker.iconPath != newAvatar) {
        marker.iconPath = newAvatar
        marker.latitude = newLat
        marker.longitude = newLng
        translating = true
        this.setData(
          {
            markers: this.data.markers,
          },
          endTranslation
        )

        return
      }

      if (marker.latitude !== newLat || marker.longitude !== newLng) {
        translating = true
        map.translateMarker({
          markerId: marker.id,
          destination: {
            latitude: newLat,
            longitude: newLng,
          },
          autoRotate: false,
          rotate: 0,
          duration: 900,
          animationEnd: endTranslation,
        })
      }
    })
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
        const carId = '62506608e4b8552ae0fa4ee0'
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
