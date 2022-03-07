const centsPerSec = 0.7

function formatDuration(sec: number) {
  const padStr = (n: number) => {
    return n.toFixed(0).padStart(2, "0")
  }
  const h = Math.floor(sec / 3600)
  sec -= h * 3600
  const m = Math.floor(sec / 60)
  sec -= m * 60
  const s = Math.floor(sec)

  return `${padStr(h)}:${padStr(m)}:${padStr(s)}`
}

function formatFee(cents: number) {
  return `${(cents / 100).toFixed(2)}元`
}

Page({
  timer: undefined as number | undefined,
  data: {
    location: {
      latitude: 32.92,
      longitude: 118.46,
    },
    scale: 14,
    elapsed: "00:00:00",
    fee: "0.00元",
  },
  onLoad() {
    this.setupLocationUpdator()
    this.setupTimer()
  },
  onUnload() {
    wx.stopLocationUpdate()
    if (this.timer) {
      clearInterval(this.timer)
    }
  },
  setupLocationUpdator() {
    wx.startLocationUpdate({
      fail: console.error,
    })

    wx.onLocationChange((loc) => {
      console.log(loc)
      this.setData({
        location: {
          latitude: loc.latitude,
          longitude: loc.longitude,
        },
      })
    })
  },
  setupTimer() {
    let elapsedSec = 0
    let cents = 0
    this.timer = setInterval(() => {
      elapsedSec += 1
      cents += centsPerSec
      this.setData({
        elapsed: formatDuration(elapsedSec),
        fee: formatFee(cents),
      })
    }, 1000)
  },
})
