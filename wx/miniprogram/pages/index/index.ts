Page({
  isPageShowing: false,
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
        iconPath: "/images/car.png",
        id: 0,
        latitude: 23.099994,
        longitude: 113.32452,
        width: 50,
        height: 50,
      },
      {
        iconPath: "/images/car.png",
        id: 0,
        latitude: 23.099994,
        longitude: 114.32452,
        width: 50,
        height: 50,
      },
    ],
  },
  onMyLocationTap() {
    wx.getLocation({
      type: "gcj02",
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
          icon: "none",
          title: "请前往设置页授权",
        })
      },
    })
  },
  moveCars() {
    const map = wx.createMapContext("map")

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

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad() {},

  /**
   * 生命周期函数--监听页面初次渲染完成
   */
  onReady() {},

  /**
   * 生命周期函数--监听页面显示
   */
  onShow() {
    this.isPageShowing = true
  },

  /**
   * 生命周期函数--监听页面隐藏
   */
  onHide() {
    this.isPageShowing = false
  },

  /**
   * 生命周期函数--监听页面卸载
   */
  onUnload() {},

  /**
   * 页面相关事件处理函数--监听用户下拉动作
   */
  onPullDownRefresh() {},

  /**
   * 页面上拉触底事件的处理函数
   */
  onReachBottom() {},

  /**
   * 用户点击右上角分享
   */
  onShareAppMessage() {},
})
