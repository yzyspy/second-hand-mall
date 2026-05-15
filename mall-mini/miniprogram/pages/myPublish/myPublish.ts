// pages/myPublish/myPublish.ts
import { get, post } from '../../utils/request'

interface ProductItem {
  id: number
  title: string
  price: number
  images: string
  status: number
  create_time: string
  firstImage: string
  statusText: string
  statusClass: string
}

interface MyPublishData {
  products: ProductItem[]
  total: number
  page: number
  pageSize: number
  loading: boolean
  hasMore: boolean
}

const STATUS_MAP: Record<number, { text: string; cls: string }> = {
  0: { text: '在售', cls: 'status-on-sale' },
  1: { text: '已售出', cls: 'status-sold' },
  2: { text: '已下架', cls: 'status-off' }
}

function processProducts(list: any[]): ProductItem[] {
  return list.map(item => ({
    ...item,
    firstImage: item.images ? item.images.split(',')[0] : '',
    statusText: STATUS_MAP[item.status]?.text ?? '未知',
    statusClass: STATUS_MAP[item.status]?.cls ?? ''
  }))
}

Page<MyPublishData, WechatMiniprogram.IAnyObject>({
  data: {
    products: [],
    total: 0,
    page: 1,
    pageSize: 10,
    loading: false,
    hasMore: false
  },

  onLoad() {
    this.loadProducts()
  },

  onShow() {
    this.setData({ products: [], page: 1, hasMore: false })
    this.loadProducts()
  },

  onPullDownRefresh() {
    this.setData({ products: [], page: 1, hasMore: false })
    this.loadProducts().finally(() => wx.stopPullDownRefresh())
  },

  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.loadMore()
    }
  },

  async loadProducts() {
    if (this.data.loading) return
    this.setData({ loading: true })
    try {
      const res = await get('/api/product/mine', { page: 1, page_size: this.data.pageSize })
      const { list, total } = res.data
      const products = processProducts(list || [])
      this.setData({
        products,
        total,
        page: 1,
        hasMore: products.length < total
      })
    } catch (err) {
      console.error('加载失败', err)
    } finally {
      this.setData({ loading: false })
    }
  },

  async loadMore() {
    if (this.data.loading) return
    const nextPage = this.data.page + 1
    this.setData({ loading: true })
    try {
      const res = await get('/api/product/mine', { page: nextPage, page_size: this.data.pageSize })
      const { list } = res.data
      const more = processProducts(list || [])
      const all = [...this.data.products, ...more]
      this.setData({
        products: all,
        page: nextPage,
        hasMore: all.length < this.data.total
      })
    } catch (err) {
      console.error('加载更多失败', err)
    } finally {
      this.setData({ loading: false })
    }
  },

  onEditTap(e: WechatMiniprogram.TouchEvent) {
    const { id } = e.currentTarget.dataset
    wx.navigateTo({ url: `/pages/productEdit/productEdit?id=${id}` })
  },

  onMarkSold(e: WechatMiniprogram.TouchEvent) {
    const { id } = e.currentTarget.dataset
    wx.showModal({
      title: '确认',
      content: '确认将此商品标记为已售出？',
      success: (res) => {
        if (!res.confirm) return
        post('/api/product/change-status', { id, status: 1 })
          .then(() => {
            wx.showToast({ title: '已标记为售出', icon: 'success' })
            this.setData({ products: [], page: 1 })
            this.loadProducts()
          })
          .catch(() => {})
      }
    })
  },

  onDelist(e: WechatMiniprogram.TouchEvent) {
    const { id } = e.currentTarget.dataset
    wx.showModal({
      title: '确认下架',
      content: '下架后商品将不再展示，确认下架？',
      success: (res) => {
        if (!res.confirm) return
        post('/api/product/change-status', { id, status: 2 })
          .then(() => {
            wx.showToast({ title: '已下架', icon: 'success' })
            this.setData({ products: [], page: 1 })
            this.loadProducts()
          })
          .catch(() => {})
      }
    })
  },

  onDelete(e: WechatMiniprogram.TouchEvent) {
    const { id } = e.currentTarget.dataset
    wx.showModal({
      title: '确认删除',
      content: '删除后不可恢复，确认删除此商品？',
      confirmColor: '#ff4d4f',
      success: (res) => {
        if (!res.confirm) return
        post('/api/product/change-status', { id, status: 2 })
          .then(() => {
            wx.showToast({ title: '已删除', icon: 'success' })
            this.setData({ products: [], page: 1 })
            this.loadProducts()
          })
          .catch(() => {})
      }
    })
  }
})
