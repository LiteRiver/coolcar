Page({
  data: {
    licenseNo: "",
    name: "",
    genderIndex: 0,
    genders: ["未知", "男", "女", "其他"],
    driverLicenseUrl: "",
    // driverLicenseUrl: "/images/sedan.png",
    dateOfBirth: "1990-01-01",
    state: "UNSUBMITTED" as "UNSUBMITTED" | "PENDING" | "VERIFIED",
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
              licenseNo: "123412341234",
              name: "CLIVE ZHANG",
              genderIndex: 1,
              dateOfBirth: "1983-06-01",
            })
          }, 1000)
        }
      },
    })
  },
  onGenderChanged(e: any) {
    this.setData({
      genderIndex: e.detail.value,
    })
  },
  onDateOfBirthChanged(e: any) {
    this.setData({
      dateOfBirth: e.detail.value,
    })
  },
  onSubmit() {
    // TODO: submit the form to server
    this.setData({
      state: "PENDING",
    })

    setTimeout(() => {
      this.onVerified()
    }, 3000)
  },
  onResubmit() {
    this.setData({
      state: "UNSUBMITTED",
      driverLicenseUrl: "",
    })
  },
  onVerified() {
    this.setData({
      state: "VERIFIED",
    })
    wx.redirectTo({
      url: "/pages/unlock/unlock",
    })
  },
})
