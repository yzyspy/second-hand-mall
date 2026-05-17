import { get } from '../../utils/request'

interface ProductItem {
  id: number
  title: string
  price: number
  images: string
  location: string
  status: number
  seller: string
  create_time: string
}

interface SearchData {
  keyword: string
  sort: string
  status: number | null
  products: ProductItem[]
  loading: boolean
  page: number
  hasMore: boolean
  sortOptions: string[]
  statusOptions: string[]
  sortIndex: number
  statusIndex: number
}

Page<SearchData, WechatMiniprogram.IAnyObject>({
  data: {
    keyword: '',
    sort: 'time_desc',
    status: null,
    products: [],
    loading: false,
    page: 1,
    hasMore: true,
    sortOptions: ['最新发布', '最早发布'],
    statusOptions: ['全部', '在售', '已售出'],
    sortIndex: 0,
    statusIndex: 0
  },

  onLoad(options: Record<string, string>) {
    if (options.keyword) {
      this.setData({ keyword: options.keyword })
      this.search()
    }
  },

  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.search()
    }
  },

  onKeywordInput(e: WechatMiniprogram.Input) {
    this.setData({ keyword: e.detail.value })
  },

  onSearch() {
    this.setData({ page: 1, hasMore: true, products: [] })
    this.search()
  },

  onSortChange(e: WechatMiniprogram.PickerChange) {
    const index = Number(e.detail.value)
    const sortMap = ['time_desc', 'time_asc']
    this.setData({ sortIndex: index, sort: sortMap[index], page: 1, hasMore: true, products: [] })
    this.search()
  },

  onStatusChange(e: WechatMiniprogram.PickerChange) {
    const index = Number(e.detail.value)
    const statusMap: (number | null)[] = [null, 0, 1]
    this.setData({ statusIndex: index, status: statusMap[index], page: 1, hasMore: true, products: [] })
    this.search()
  },

  async search() {
    if (this.data.loading) return
    this.setData({ loading: true })

    try {
      const params: Record<string, string | number> = {
        keyword: this.data.keyword,
        sort: this.data.sort,
        page: this.data.page,
        page_size: 10
      }
      if (this.data.status !== null) {
        params.status = this.data.status
      }
      const res = await get<{
        list: ProductItem[]
        total: number
        page: number
        page_size: number
      }>('/api/product/search', params)

      if (res.code === 0 && res.data) {
        const newList = res.data.list || []
        this.setData({
          products: this.data.page === 1 ? newList : [...this.data.products, ...newList],
          hasMore: this.data.products.length + newList.length < res.data.total,
          page: this.data.page + 1
        })
      }
    } catch {
      // 请求失败，request.ts 已处理 toast 提示
    } finally {
      this.setData({ loading: false })
    }
  },

  goToDetail(e: WechatMiniprogram.TouchEvent) {
    const { id } = e.currentTarget.dataset
    wx.navigateTo({ url: `/pages/detail/detail?id=${id}` })
  }
})
