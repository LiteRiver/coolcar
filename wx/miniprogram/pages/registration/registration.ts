import { ProfileService } from '../../services/profile'
import { rental } from '../../services/proto-gen/rental/rental-pb'
import { Coolcar } from '../../services/request'
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

    const photoRes = await ProfileService.getPhoto()
    if (photoRes.url) {
      this.setData({
        driverLicenseUrl: photoRes.url || '',
      })
    }
    const p = await ProfileService.get()
    this.renderProfile(p)
  },
  onUnload() {
    this.clearProfileRefresher()
  },
  renderProfile(profile: rental.v1.IProfile) {
    this.renderIdentity(profile.identity!)
    this.setData({
      state: rental.v1.IdentityStatus[profile.identityStatus || 0],
    })
  },
  renderIdentity(identity?: rental.v1.IIdentity) {
    this.setData({
      licenseNo: identity?.licenseNumber || '',
      name: identity?.name || '',
      genderIndex: identity?.gender || 0,
      dateOfBirth: formatDate(identity?.dateOfBirthMs || 0),
    })
  },
  onUploadLicense() {
    wx.chooseImage({
      success: async (res) => {
        if (res.tempFilePaths.length === 0) {
          return
        }

        this.setData({
          driverLicenseUrl: res.tempFilePaths[0],
        })

        const photoRes = await ProfileService.createPhoto()
        if (!photoRes.uploadUrl) {
          return
        }

        await Coolcar.uploadFile({
          localPath: res.tempFilePaths[0],
          url: photoRes.uploadUrl,
        })

        const identity = await ProfileService.completePhoto()
        this.renderIdentity(identity)
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
  onResubmit() {
    ProfileService.clear().then((p) => this.renderProfile(p))
    ProfileService.clearPhoto().then(() => {
      this.setData({
        driverLicenseUrl: '',
      })
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
