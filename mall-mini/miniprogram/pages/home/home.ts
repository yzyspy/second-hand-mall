// pages/home/home.ts

interface ProductItem {
  id: number
  title: string
  price: number
  location: string
  images: string[]
  seller: string
  createTime: string
}

interface HomeData {
  products: ProductItem[]
  loading: boolean
  page: number
  hasMore: boolean
}

Page<HomeData, WechatMiniprogram.IAnyObject>({
  data: {
    products: [],
    loading: false,
    page: 1,
    hasMore: true
  },

  onLoad() {
    this.loadProducts()
  },

  onShow() {
    // 页面显示时刷新数据
  },

  onPullDownRefresh() {
    this.setData({ page: 1, hasMore: true })
    this.loadProducts().then(() => {
      wx.stopPullDownRefresh()
    })
  },

  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.loadProducts()
    }
  },

  // 加载商品列表
  async loadProducts() {
    if (this.data.loading) return

    this.setData({ loading: true })

    // 模拟数据，实际应调用后端API
    const mockProducts: ProductItem[] = [
      {
        id: 1,
        title: 'iPhone 13 Pro 256G 暗夜绿',
        price: 4599,
        location: '北京朝阳区',
        images: ['https://via.placeholder.com/200x200'],
        seller: '小明',
        createTime: '2024-01-15'
      },
      {
        id: 2,
        title: 'MacBook Pro 14寸 M2 Pro',
        price: 12999,
        location: '上海浦东新区',
        images: ['https://via.placeholder.com/200x200'],
        seller: '科技达人',
        createTime: '2024-01-14'
      },
      {
        id: 3,
        title: 'Sony WH-1000XM5 降噪耳机',
        price: 1899,
        location: '深圳南山区',
        images: ['https://via.placeholder.com/200x200'],
        seller: '数码控',
        createTime: '2024-01-13'
      }
    ]

    // 模拟网络延迟
    setTimeout(() => {
      this.setData({
        products: this.data.page === 1 ? mockProducts : [...this.data.products, ...mockProducts],
        loading: false,
        page: this.data.page + 1,
        hasMore: this.data.page < 3
      })
    }, 500)
  },

  // 跳转到商品详情
  goToDetail(e: WechatMiniprogram.TouchEvent) {
    const { id } = e.currentTarget.dataset
    wx.navigateTo({
      url: `/pages/detail/detail?id=${id}`
    })
  },

  // 搜索
  onSearch() {
    wx.navigateTo({
      url: '/pages/search/search'
    })
  }
})
