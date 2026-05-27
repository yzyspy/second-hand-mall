import { request } from './utils/request'

App<IAppOption>({
  globalData: {
    userInfo: null,
    token: '',
    baseUrl: 'http://localhost:8080'
  },

  onLaunch() {
    const token = wx.getStorageSync('token')
    if (token) {
      this.globalData.token = token
    }
  },

  onShow() {
    const token = wx.getStorageSync('token')
    if (!token) return
    request<{ count: number }>({ url: '/api/chat/unread-count' }).then(response => {
      if (response.data && response.data.count > 0) {
        wx.setTabBarBadge({ index: 2, text: String(response.data.count) })
      } else {
        wx.removeTabBarBadge({ index: 2 })
      }
    }).catch(() => {/* ignore if not logged in */})
  }
})
