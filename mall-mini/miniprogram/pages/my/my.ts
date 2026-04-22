// pages/my/my.ts

interface MenuItem {
  icon: string
  title: string
  badge?: string
  action: string
}

interface MyData {
  userInfo: {
    avatarUrl: string
    nickName: string
  } | null
  isLoggedIn: boolean
  menuItems: MenuItem[]
  publishCount: number
  soldCount: number
  boughtCount: number
}

Page<MyData, WechatMiniprogram.IAnyObject>({
  data: {
    userInfo: null,
    isLoggedIn: false,
    menuItems: [
      { icon: '📦', title: '我发布的', action: 'myPublish' },
      { icon: '💰', title: '我卖出的', action: 'mySold' },
      { icon: '🛒', title: '我买到的', action: 'myBought' },
      { icon: '❤️', title: '我的收藏', action: 'myFavorite' },
      { icon: '📍', title: '收货地址', action: 'myAddress' },
      { icon: '⚙️', title: '设置', action: 'settings' },
      { icon: '📞', title: '联系客服', action: 'contactService' },
      { icon: 'ℹ️', title: '关于我们', action: 'about' }
    ],
    publishCount: 0,
    soldCount: 0,
    boughtCount: 0
  },

  onLoad() {
    this.checkLoginStatus()
  },

  onShow() {
    this.checkLoginStatus()
  },

  // 检查登录状态
  checkLoginStatus() {
    const token = wx.getStorageSync('token')
    const userInfo = wx.getStorageSync('userInfo')

    this.setData({
      isLoggedIn: !!token,
      userInfo: userInfo || null
    })
  },

  // 登录
  async handleLogin() {
    try {
      // 获取用户信息
      const { userInfo } = await wx.getUserProfile({
        desc: '用于完善用户资料'
      })

      // 调用后端登录接口
      // const result = await post('/user/login', { ... })

      // 模拟登录成功
      wx.setStorageSync('userInfo', userInfo)
      wx.setStorageSync('token', 'mock_token')

      this.setData({
        userInfo,
        isLoggedIn: true
      })

      wx.showToast({ title: '登录成功', icon: 'success' })
    } catch (err) {
      console.log('登录失败', err)
    }
  },

  // 退出登录
  handleLogout() {
    wx.showModal({
      title: '提示',
      content: '确定要退出登录吗？',
      success: (res) => {
        if (res.confirm) {
          wx.removeStorageSync('token')
          wx.removeStorageSync('userInfo')
          this.setData({
            userInfo: null,
            isLoggedIn: false
          })
          wx.showToast({ title: '已退出登录', icon: 'success' })
        }
      }
    })
  },

  // 编辑资料
  editProfile() {
    wx.navigateTo({
      url: '/pages/profile/profile'
    })
  },

  // 菜单点击
  onMenuTap(e: WechatMiniprogram.TouchEvent) {
    const { action } = e.currentTarget.dataset

    if (!this.data.isLoggedIn && ['myPublish', 'mySold', 'myBought', 'myFavorite', 'myAddress'].includes(action)) {
      wx.showToast({ title: '请先登录', icon: 'none' })
      return
    }

    switch (action) {
      case 'myPublish':
        wx.navigateTo({ url: '/pages/myPublish/myPublish' })
        break
      case 'mySold':
        wx.navigateTo({ url: '/pages/mySold/mySold' })
        break
      case 'myBought':
        wx.navigateTo({ url: '/pages/myBought/myBought' })
        break
      case 'myFavorite':
        wx.navigateTo({ url: '/pages/myFavorite/myFavorite' })
        break
      case 'myAddress':
        wx.navigateTo({ url: '/pages/myAddress/myAddress' })
        break
      case 'settings':
        wx.navigateTo({ url: '/pages/settings/settings' })
        break
      case 'contactService':
        wx.makePhoneCall({ phoneNumber: '400-123-4567' })
        break
      case 'about':
        wx.navigateTo({ url: '/pages/about/about' })
        break
    }
  }
})
