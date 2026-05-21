import { get } from '../../utils/request'
import regionsData from '../../data/china-regions'

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
  selectedCategory: string
  selectedProvince: string
  selectedCity: string
  selectedDistrict: string
  showCategoryPanel: boolean
  showRegionPanel: boolean
  regionStep: number
  regionProvinceIndex: number
  regionCityIndex: number
  regionCities: string[]
  regionDistricts: string[]
  provinces: string[]
  categories: string[]
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
    selectedCategory: '',
    selectedProvince: '',
    selectedCity: '',
    selectedDistrict: '',
    showCategoryPanel: false,
    showRegionPanel: false,
    regionStep: 0,
    regionProvinceIndex: 0,
    regionCityIndex: 0,
    regionCities: [],
    regionDistricts: [],
    provinces: regionsData.map(p => p.name),
    categories: ['电子产品', '服装鞋帽', '图书文具', '生活用品', '数码配件', '其他'],
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
        page_size: 10,
      }
      if (this.data.status !== null) {
        params.status = this.data.status
      }
      if (this.data.selectedCategory) params.category = this.data.selectedCategory
      if (this.data.selectedProvince) params.province = this.data.selectedProvince
      if (this.data.selectedCity)     params.city     = this.data.selectedCity
      if (this.data.selectedDistrict) params.district = this.data.selectedDistrict

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

  onOpenCategoryPanel() {
    this.setData({ showCategoryPanel: true })
  },

  onCloseCategoryPanel() {
    this.setData({ showCategoryPanel: false })
  },

  onSelectCategory(e: WechatMiniprogram.TouchEvent) {
    const value: string = e.currentTarget.dataset.value
    this.setData({ selectedCategory: value, showCategoryPanel: false, page: 1, hasMore: true, products: [] })
    this.search()
  },

  onClearCategory(_e: WechatMiniprogram.TouchEvent) {
    this.setData({ selectedCategory: '', page: 1, hasMore: true, products: [] })
    this.search()
  },

  onOpenRegionPanel() {
    this.setData({ showRegionPanel: true, regionStep: 0 })
  },

  onCloseRegionPanel() {
    this.setData({ showRegionPanel: false })
  },

  onSelectProvince(e: WechatMiniprogram.TouchEvent) {
    const pi: number = e.currentTarget.dataset.index
    const province = regionsData[pi]
    this.setData({
      selectedProvince: province.name,
      selectedCity: '',
      selectedDistrict: '',
      regionProvinceIndex: pi,
      regionCities: province.children.map((c: any) => c.name),
      regionStep: 1,
    })
  },

  onSelectCity(e: WechatMiniprogram.TouchEvent) {
    const ci: number = e.currentTarget.dataset.index
    const pi = this.data.regionProvinceIndex
    const city = regionsData[pi].children[ci]
    this.setData({
      selectedCity: city.name,
      selectedDistrict: '',
      regionCityIndex: ci,
      regionDistricts: city.children as string[],
      regionStep: 2,
    })
  },

  onSelectDistrict(e: WechatMiniprogram.TouchEvent) {
    const district: string = e.currentTarget.dataset.district
    this.setData({ selectedDistrict: district, showRegionPanel: false, page: 1, hasMore: true, products: [] })
    this.search()
  },

  onConfirmRegion() {
    this.setData({ showRegionPanel: false, page: 1, hasMore: true, products: [] })
    this.search()
  },

  onClearRegion(_e: WechatMiniprogram.TouchEvent) {
    this.setData({ selectedProvince: '', selectedCity: '', selectedDistrict: '', page: 1, hasMore: true, products: [] })
    this.search()
  },

  goToDetail(e: WechatMiniprogram.TouchEvent) {
    const { id } = e.currentTarget.dataset
    wx.navigateTo({ url: `/pages/detail/detail?id=${id}` })
  }
})
