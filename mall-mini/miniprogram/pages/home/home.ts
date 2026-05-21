// pages/home/home.ts

import { get } from '../../utils/request'
import regionsData from '../../data/china-regions'

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
}

Page<HomeData, WechatMiniprogram.IAnyObject>({
  data: {
    products: [],
    loading: false,
    page: 1,
    hasMore: true,
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

  async loadProducts() {
    if (this.data.loading) return

    this.setData({ loading: true })

    try {
      const params: Record<string, string | number> = {
        page: this.data.page,
        page_size: 10,
        status: 0,
      }
      if (this.data.selectedCategory) params.category = this.data.selectedCategory
      if (this.data.selectedProvince) params.province = this.data.selectedProvince
      if (this.data.selectedCity)     params.city     = this.data.selectedCity
      if (this.data.selectedDistrict) params.district = this.data.selectedDistrict

      const response = await get<{ list: any[], total: number, page: number, page_size: number }>(
        '/api/product/search',
        params
      )

      if (response.code === 0 && response.data) {
        const { list } = response.data

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

  goToDetail(e: WechatMiniprogram.TouchEvent) {
    const { id } = e.currentTarget.dataset
    wx.navigateTo({
      url: `/pages/detail/detail?id=${id}`
    })
  },

  onSearch() {
    wx.navigateTo({
      url: '/pages/search/search'
    })
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
    this.loadProducts()
  },

  onClearCategory(_e: WechatMiniprogram.TouchEvent) {
    this.setData({ selectedCategory: '', page: 1, hasMore: true, products: [] })
    this.loadProducts()
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
    this.loadProducts()
  },

  onConfirmRegion() {
    this.setData({ showRegionPanel: false, page: 1, hasMore: true, products: [] })
    this.loadProducts()
  },

  onClearRegion(_e: WechatMiniprogram.TouchEvent) {
    this.setData({ selectedProvince: '', selectedCity: '', selectedDistrict: '', page: 1, hasMore: true, products: [] })
    this.loadProducts()
  },
})
