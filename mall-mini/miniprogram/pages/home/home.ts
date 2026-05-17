// pages/home/home.ts

import { get } from '../../utils/request'

interface ProductItem {
  id: number
  title: string
  price: number
  location: string
  images: string[]
  seller: string
  avatar: string
  buy_uid: number
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
    this.setData({ page: 1, hasMore: true, products: [] })
    this.loadProducts()
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

    try {
      const response = await get<{ list: any[], total: number, page: number, page_size: number }>(
        '/api/product/search',
        {
          page: this.data.page,
          page_size: 10,
          status: 0 // 只获取在售商品
        }
      )

      if (response.code === 0 && response.data) {
        const { list } = response.data

        // 转换数据格式
        const products: ProductItem[] = list.map((item: any) => ({
          id: item.id,
          title: item.title,
          price: item.price,
          location: item.location,
          images: item.images ? item.images.split(',').filter((img: string) => img) : [],
          seller: item.seller || '匿名用户',
          avatar: item.avatar || '',
          buy_uid: item.buy_uid || 0,
          createTime: item.create_time || ''
        }))

        this.setData({
          products: this.data.page === 1 ? products : [...this.data.products, ...products],
          loading: false,
          page: this.data.page + 1,
          hasMore: list.length >= 10
        })
      } else {
        this.setData({ loading: false })
      }
    } catch (err) {
      console.error('加载商品列表失败:', err)
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
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
