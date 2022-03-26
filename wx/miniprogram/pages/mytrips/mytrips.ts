interface Trip {
  id: string
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
  data: Trip
}

interface NavItem {
  mainId: string
  navId: string
  label: string
}

Page({
  data: {
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
  },
  onLoad() {
    this.populateTrips()
  },
  onReady() {
    wx.createSelectorQuery()
      .select('#heading')
      .boundingClientRect((rect) => {
        const tripsHeight = wx.getSystemInfoSync().windowHeight - rect.height
        this.setData({
          tripsHeight,
          navCount: tripsHeight / 50,
        })
      })
      .exec()
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
  populateTrips() {
    const mainItems: MainItem[] = []
    const navItems: NavItem[] = []
    let activeNavItem = ''
    for (let i = 0; i < 100; i++) {
      const mainId = `main-${i}`
      const navId = `nav-${i}`
      const tripId = (10001 + i).toString()
      mainItems.push({
        mainId,
        navId,
        data: {
          id: tripId,
          start: '东方明珠',
          end: '迪斯尼',
          distance: '27.0公里',
          duration: '0时44分',
          fee: '128.0元',
          status: '已完成',
        },
      })

      navItems.push({
        mainId,
        navId,
        label: tripId,
      })

      if (i === 0) {
        activeNavItem = navId
      }
    }
    this.setData({
      mainItems,
      navItems,
      activeNavItem,
    })
  },
  onSwiperChanged(e: any) {
    console.log(e.detail)
  },
})
