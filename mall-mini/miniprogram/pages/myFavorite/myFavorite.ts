import { get, post } from '../../utils/request'

interface FavoriteItem {
  id: number
  title: string
  price: number
  images: string
  location: string
  seller: string
  avatar: string
  firstImage: string
}

interface MyFavoriteData {
  items: FavoriteItem[]
  total: number
  page: number
  pageSize: number
  loading: boolean
  hasMore: boolean
}

function processItems(list: any[]): FavoriteItem[] {
  return list.map(item => ({
    ...item,
    firstImage: item.images ? item.images.split(',')[0] : ''
  }))
}

Page<MyFavoriteData, WechatMiniprogram.IAnyObject>({
  data: {
    items: [],
    total: 0,
    page: 1,
    pageSize: 10,
    loading: false,
    hasMore: false
  },

  onLoad() {
    this.loadItems()
  },

  onShow() {
    this.setData({ items: [], page: 1, hasMore: false })
    this.loadItems()
  },

  onPullDownRefresh() {
    this.setData({ items: [], page: 1, hasMore: false })
    this.loadItems().finally(() => wx.stopPullDownRefresh())
  },

  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.loadMore()
    }
  },

  async loadItems() {
    if (this.data.loading) return
    this.setData({ loading: true })
    try {
      const res = await get('/api/favorite/list', { page: 1, page_size: this.data.pageSize })
      const { list, total } = res.data
      const items = processItems(list || [])
      this.setData({ items, total, page: 1, hasMore: items.length < total })
    } catch (err) {
      console.error('加载收藏失败', err)
    } finally {
      this.setData({ loading: false })
    }
  },

  async loadMore() {
    if (this.data.loading) return
    const nextPage = this.data.page + 1
    this.setData({ loading: true })
    try {
      const res = await get('/api/favorite/list', { page: nextPage, page_size: this.data.pageSize })
      const { list } = res.data
      const more = processItems(list || [])
      const all = [...this.data.items, ...more]
      this.setData({ items: all, page: nextPage, hasMore: all.length < this.data.total })
    } catch (err) {
      console.error('加载更多失败', err)
    } finally {
      this.setData({ loading: false })
    }
  },

  goToDetail(e: WechatMiniprogram.TouchEvent) {
    const { id } = e.currentTarget.dataset
    wx.navigateTo({ url: `/pages/detail/detail?id=${id}` })
  },

  async onUnfavorite(e: WechatMiniprogram.TouchEvent) {
    const { id } = e.currentTarget.dataset
    try {
      await post('/api/favorite/toggle', { product_id: id })
      const items = this.data.items.filter(item => item.id !== id)
      this.setData({ items, total: this.data.total - 1 })
    } catch (err) {
      console.error('取消收藏失败', err)
    }
  }
})
