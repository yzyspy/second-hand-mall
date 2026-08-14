import { defineStore } from 'pinia'
import { searchProducts } from '../api/product'
import { buildSearchParams } from '../utils/searchParams'

export const useHomeStore = defineStore('home', {
  state: () => ({
    products: [],
    page: 1,
    hasMore: true,
    loading: false,
    category: '',
    province: '',
    city: '',
    district: ''
  }),
  actions: {
    setCategory(category) {
      this.category = category
      this.loadProducts(true)
    },
    setRegion({ province, city, district }) {
      this.province = province
      this.city = city
      this.district = district
      this.loadProducts(true)
    },
    async loadProducts(reset) {
      if (this.loading) return
      if (reset) {
        this.page = 1
        this.hasMore = true
        this.products = []
      }
      this.loading = true
      try {
        const params = buildSearchParams({
          page: this.page,
          category: this.category,
          province: this.province,
          city: this.city,
          district: this.district
        })
        const res = await searchProducts(params)
        const list = res.data?.list || []
        const items = list.map((item) => ({
          id: item.id,
          title: item.title,
          price: item.price,
          location: item.location,
          images: item.images ? item.images.split(',').filter(Boolean) : [],
          seller: item.seller || '匿名用户',
          avatar: item.avatar || ''
        }))
        this.products = reset ? items : [...this.products, ...items]
        this.hasMore = list.length >= 10
        this.page += 1
      } finally {
        this.loading = false
      }
    }
  }
})
