// app.ts
App<IAppOption>({
  globalData: {
    userInfo: null,
    token: '',
    baseUrl: 'http://localhost:8080'
  },

  onLaunch() {
    // 检查登录状态
    const token = wx.getStorageSync('token')
    if (token) {
      this.globalData.token = token
    }
  }
})
