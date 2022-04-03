import { ProfileService } from '../../services/profile'
import { rental } from '../../services/proto-gen/rental/rental-pb'
import { formats } from '../../utils/formats'
import { routing } from '../../utils/routing'

function formatDate(ms: number) {
  const dt = new Date(ms)
  const y = formats.padStartZero(dt.getFullYear(), 2)
  const m = formats.padStartZero(dt.getMonth() - 1, 2)
  const d = formats.padStartZero(dt.getDate(), 2)

  return `${y}-${m}-${d}`
}

Page({
  redirectURL: '',
  profileRefresher: 0,
  data: {
    licenseNo: '',
    name: '',
    genderIndex: 0,
    genders: ['未知', '男', '女'],
    driverLicenseUrl: '',
    // driverLicenseUrl: "/images/sedan.png",
    dateOfBirth: '1990-01-01',
    state: rental.v1.IdentityStatus[rental.v1.IdentityStatus.UNSUBMITTED],
  },
  async onLoad(opts: Record<'redirect', string>) {
    const registrationOpts: routing.RegistrationOpts = opts
    if (registrationOpts.redirect) {
      this.redirectURL = decodeURIComponent(registrationOpts.redirect)
    }
    const profile = await ProfileService.get()
    this.renderProfile(profile)
  },
  onUnload() {
    this.clearProfileRefresher()
  },
  renderProfile(profile: rental.v1.IProfile) {
    this.setData({
      licenseNo: profile.identity?.licenseNumber || '',
      name: profile.identity?.name || '',
      genderIndex: profile.identity?.gender || 0,
      dateOfBirth: formatDate(profile.identity?.dateOfBirthMs || 0),
      state: rental.v1.IdentityStatus[profile.identityStatus || 0],
    })
  },
  onUploadLicense() {
    wx.chooseImage({
      success: (res) => {
        if (res.tempFilePaths.length > 0) {
          this.setData({
            driverLicenseUrl: res.tempFilePaths[0],
          })
          // TODO: upload license image
          setTimeout(() => {
            this.setData({
              licenseNo: '123412341234',
              name: 'CLIVE ZHANG',
              genderIndex: 1,
              dateOfBirth: '1983-06-01',
            })
          }, 1000)
        }
      },
    })
  },
  onGenderChanged(e: any) {
    this.setData({
      genderIndex: parseInt(e.detail.value),
    })
  },
  onDateOfBirthChanged(e: any) {
    this.setData({
      dateOfBirth: e.detail.value,
    })
  },
  async onSubmit() {
    const profile = await ProfileService.submit({
      licenseNumber: this.data.licenseNo,
      name: this.data.name,
      gender: this.data.genderIndex,
      dateOfBirthMs: Date.parse(this.data.dateOfBirth),
    })

    this.renderProfile(profile)
    this.scheduleProfileRefresher()
  },
  scheduleProfileRefresher() {
    this.profileRefresher = setInterval(async () => {
      const profile = await ProfileService.get()
      this.renderProfile(profile)
      if (profile.identityStatus !== rental.v1.IdentityStatus.PENDING) {
        this.clearProfileRefresher()
      }

      if (profile.identityStatus === rental.v1.IdentityStatus.VERIFIED) {
        this.onVerified()
      }
    }, 1000)
  },
  async clearProfileRefresher() {
    if (this.profileRefresher) {
      clearInterval(this.profileRefresher)
      this.profileRefresher = 0
    }
  },
  async onResubmit() {
    const profile = await ProfileService.clear()
    this.renderProfile(profile)
    this.setData({
      driverLicenseUrl: '',
    })
  },
  onVerified() {
    if (this.redirectURL) {
      wx.redirectTo({
        url: this.redirectURL,
      })
    }
  },
})
