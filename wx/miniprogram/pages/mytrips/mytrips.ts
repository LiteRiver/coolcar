import { IAppOption } from '../../app-option'
import { ProfileService } from '../../services/profile'
import { rental } from '../../services/proto-gen/rental/rental-pb'
import { TripService } from '../../services/trip'
import { consts } from '../../utils/consts'
import { formats } from '../../utils/formats'
import { routing } from '../../utils/routing'
import { wxapi } from '../../utils/wxapi'

interface Trip {
  id: string
  shortId: string
  start: string
  end: string
  duration: string
  fee: string
  distance: string
  status: string
}

interface MainItem {
  mainId: string
  navId: string
  navScrollId: string
  data: Trip
}

interface NavItem {
  mainId: string
  navId: string
  label: string
}

interface MainItemState {
  mainId: string
  top: number
  dataset: {
    navId: string
    navScrollId: string
  }
}

const tripStatusMap = new Map([
  [rental.v1.TripStatus.IN_PROGRESS, '进行中'],
  [rental.v1.TripStatus.FINISHED, '已完成'],
])

const app = getApp<IAppOption>()

const IdentityStatusMapping = new Map([
  [rental.v1.IdentityStatus.UNSUBMITTED, '未认证'],
  [rental.v1.IdentityStatus.PENDING, '认证中'],
  [rental.v1.IdentityStatus.VERIFIED, '已认证'],
])

Page({
  scrollStates: {
    mainItems: [] as MainItemState[],
  },
  layoutResolver: undefined as ((value: unknown) => void) | undefined,
  data: {
    avatarUrl: '',
    imageUrls: [
      'https://img2.mukewang.com/622211a20001a95817920764.jpg',
      'https://img1.mukewang.com/62256a5d0001eb3e17920764.jpg',
      'https://img4.mukewang.com/62205ba4000134df17920764.jpg',
      'https://img2.mukewang.com/621ca25c00011f4017920764.jpg',
    ],
    current: 0,
    navCount: 0,
    activeNavItem: '',
    mainItems: [] as MainItem[],
    navItems: [] as NavItem[],
    mainScroll: '',
    navScroll: '',
    identityStatus: IdentityStatusMapping.get(rental.v1.IdentityStatus.UNSUBMITTED),
  },
  onLoad() {
    const layoutReady = new Promise((resolve) => {
      this.layoutResolver = resolve
    })
    Promise.all([TripService.list(), layoutReady]).then(([res, _]) => {
      this.populateTrips(res.trips)
    })

    app.globalData.userInfo.then((userInfo) => {
      this.setData({
        avatarUrl: userInfo.avatarUrl,
      })
    })
  },
  onReady() {
    wx.createSelectorQuery()
      .select('#heading')
      .boundingClientRect((rect) => {
        const tripsHeight = wx.getSystemInfoSync().windowHeight - rect.height
        this.setData(
          {
            tripsHeight,
            navCount: tripsHeight / 50,
          },
          () => {
            if (this.layoutResolver) {
              this.layoutResolver(undefined)
            }
          }
        )
      })
      .exec()
  },
  async onShow() {
    const profile = await ProfileService.get()
    console.log(IdentityStatusMapping.get(profile.identityStatus))
    this.setData({
      identityStatus: IdentityStatusMapping.get(profile.identityStatus),
    })
  },
  onGetUserProfile() {
    wxapi.getUserProfile().then((res) => {
      app.resolveUserInfo(res.userInfo)
      wx.setStorageSync(consts.ShareLocationKey, true)
    })
  },
  onRegisterClicked() {
    wx.navigateTo({
      url: routing.registration(),
    })
  },
  onNavItemClicked(e: any) {
    const mainId: string = e.currentTarget?.dataset?.mainId
    const navId: string = e.currentTarget?.dataset?.navId
    console.log(mainId, navId)
    if (mainId && navId) {
      this.setData({
        mainScroll: mainId,
        activeNavItem: navId,
      })
    }
  },
  populateTrips(trips: rental.v1.ITripEntity[]) {
    const mainItems: MainItem[] = []
    const navItems: NavItem[] = []
    let activeNavItem = ''
    let prevNav = ''
    for (let i = 0; i < trips.length; i++) {
      const trip = trips[i]
      const mainId = `main-${i}`
      const navId = `nav-${i}`
      const tripId = trip.id!
      const shortId = '****' + tripId.substr(tripId.length - 6)
      if (!prevNav) {
        prevNav = navId
      }

      const tripData: Trip = {
        id: tripId,
        shortId,
        start: trip.trip?.start?.pointName || '未知',
        status: tripStatusMap.get(trip.trip?.status!) || '未知',
        end: '',
        distance: '',
        duration: '',
        fee: '',
      }

      const end = trip.trip?.end
      if (end) {
        tripData.end = end.pointName || '未知'
        tripData.distance = end.kmDriven?.toFixed(1) + '公里'
        tripData.fee = formats.formatFee(end.feeCent || 0)
        const dur = formats.formatDuration(
          (end.timestampSec || 0) - (trip.trip?.start?.timestampSec || 0)
        )
        tripData.duration = `${dur.hh}时${dur.mm}分`
      }

      mainItems.push({
        mainId,
        navId,
        navScrollId: prevNav,
        data: tripData,
      })

      navItems.push({
        mainId,
        navId,
        label: shortId,
      })

      if (i === 0) {
        activeNavItem = navId
      }

      prevNav = navId
    }

    for (let i = 0; i < this.data.navCount - 1; i++) {
      navItems.push({
        mainId: '',
        navId: '',
        label: '',
      })
    }
    this.setData(
      {
        mainItems,
        navItems,
        activeNavItem,
      },
      () => {
        this.prepareScrollStates()
      }
    )
  },
  prepareScrollStates() {
    wx.createSelectorQuery()
      .selectAll('.trip')
      .fields({
        id: true,
        dataset: true,
        rect: true,
      })
      .exec((res) => {
        this.scrollStates.mainItems = res[0]
      })
  },
  onMainScroll(e: any) {
    const top: number = e.currentTarget?.offsetTop + e.detail?.scrollTop
    if (top === undefined) {
      return
    }

    const activeItem = this.scrollStates.mainItems.find((i) => i.top >= top)
    if (!activeItem) {
      return
    }

    this.setData({
      activeNavItem: activeItem.dataset.navId,
      navScroll: activeItem.dataset.navScrollId,
    })
  },
})
