import { rental } from '../../services/proto-gen/rental/rental-pb'
import { TripService } from '../../services/trip'
import { formats } from '../../utils/formats'
import { routing } from '../../utils/routing'

const updateIntervalSec = 5
function getDurationStr(sec: number) {
  const ret = formats.formatDuration(sec)
  return `${ret.hh}:${ret.mm}:${ret.ss}`
}

Page({
  timer: undefined as number | undefined,
  tripId: '',
  data: {
    location: {
      latitude: 32.92,
      longitude: 118.46,
    },
    scale: 14,
    elapsed: '00:00:00',
    fee: '0.00元',
  },
  onLoad(opts: Record<'tripId', string>) {
    const drivingOpts: routing.DrivingOpts = opts
    this.tripId = drivingOpts.tripId
    this.setupLocationUpdator()
    this.setupTimer(this.tripId)
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
  async setupTimer(tripId: string) {
    const trip = await TripService.get(tripId)
    if (trip.status !== rental.v1.TripStatus.IN_PROGRESS) {
      console.error('trip not in progress')
      return
    }

    let sinceLastUpdateSec = 0
    let lastUpdateDurationSec = trip.current!.timestampSec! - trip.start!.timestampSec!
    this.setData({
      elapsed: getDurationStr(lastUpdateDurationSec),
      fee: formats.formatFee(trip.current!.feeCent!),
    })
    this.timer = setInterval(() => {
      sinceLastUpdateSec++
      if (sinceLastUpdateSec % updateIntervalSec === 0) {
        TripService.get(tripId)
          .then((trip) => {
            lastUpdateDurationSec =
              trip.current!.timestampSec! - trip.start!.timestampSec!
            this.setData({
              fee: formats.formatFee(trip.current!.feeCent!),
            })
          })
          .catch(console.error)
      }
      this.setData({
        elapsed: getDurationStr(lastUpdateDurationSec + sinceLastUpdateSec),
      })
    }, 1000)
  },
  async onEndClicked() {
    try {
      await TripService.endTrip(this.tripId)
      wx.redirectTo({
        url: routing.mytrips(),
      })
    } catch (err) {
      console.error(err)
      wx.showToast({
        title: '结束行程失败',
        icon: 'none',
      })
    }
  },
})
