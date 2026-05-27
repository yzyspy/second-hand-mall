// pages/detail/detail.ts
import { get, post } from '../../utils/request'

interface Seller {
  id: string
  name: string
  avatar: string
  rating: number
}

interface Product {
  id: number
  title: string
  description: string
  price: number
  originalPrice?: number
  images: string[]
  category: string
  condition: string
  location: string
  seller: Seller
  createdAt: string
  views: number
  contactType: string
  contactValue: string
}

Page({
  data: {
    productId: '',
    product: {
      id: 0,
      title: '',
      description: '',
      price: 0,
      originalPrice: 0,
      images: [] as string[],
      category: '',
      condition: '九成新',
      location: '',
      seller: { id: '', name: '微信用户', avatar: '', rating: 4.5 },
      createdAt: '',
      views: 0,
      contactType: '',
      contactValue: ''
    },
    isFavorite: false,
    loading: true,
    error: null as string | null
  },

  onLoad(options: { id?: string }) {
    const { id } = options
    if (id) {
      this.setData({ productId: id })
    }
    this.loadProductDetail()
  },

  onPullDownRefresh() {
    this.loadProductDetail()
  },

  onShareAppMessage() {
    const { product } = this.data
    return {
      title: product.title,
      path: `/pages/detail/detail?id=${product.id}`,
      imageUrl: product.images[0] || ''
    }
  },

  async loadProductDetail() {
    this.setData({ loading: true, error: null })

    try {
      const response = await get<any>('/api/product/detail', { id: this.data.productId })

      if (response.code === 0 && response.data) {
        const data = response.data
        const images = data.images ? data.images.split(',').filter((img: string) => img) : []

        const product: Product = {
          id: data.id,
          title: data.title,
          description: data.description || '',
          price: data.price,
          images,
          category: '二手好物',
          condition: '九成新',
          location: data.location || '',
          seller: {
            id: String(data.user_id || 0),
            name: data.seller || '微信用户',
            avatar: data.avatar || '',
            rating: 4.8
          },
          createdAt: data.create_time ? data.create_time.substring(0, 10) : '',
          views: Math.floor(Math.random() * 500) + 50,
          contactType: data.contact_type || '',
          contactValue: data.contact_value || ''
        }

        this.setData({
          product,
          isFavorite: !!data.is_favorited,
          loading: false
        })
      } else {
        this.setData({ error: response.msg || '加载失败', loading: false })
      }
    } catch (err) {
      console.error('加载商品详情失败:', err)
      this.setData({ error: '网络错误，请重试', loading: false })
    }

    wx.stopPullDownRefresh()
  },

  async toggleFavorite() {
    const token = wx.getStorageSync('token')
    if (!token) {
      wx.showToast({ title: '请先登录', icon: 'none' })
      return
    }

    try {
      const res = await post<{ is_favorited: boolean }>('/api/favorite/toggle', {
        product_id: this.data.product.id
      })
      if (res.code === 0 && res.data) {
        const isFavorite = res.data.is_favorited
        this.setData({ isFavorite })
        wx.showToast({ title: isFavorite ? '已收藏' : '已移除收藏', icon: 'success' })
      }
    } catch (err) {
      console.error('收藏操作失败:', err)
      wx.showToast({ title: '操作失败，请重试', icon: 'none' })
    }
  },

  goToChat() {
    const productId = this.data.product.id
    const sellerId = this.data.product.seller.id
    if (!productId || !sellerId) return
    wx.navigateTo({
      url: `/pages/chat/chat?product_id=${productId}&receiver_id=${sellerId}`
    })
  },

  reportProduct() {
    wx.showActionSheet({
      itemList: ['虚假信息', '骚扰信息', '违法违规'],
      success: () => {
        wx.showToast({ title: '举报成功', icon: 'success' })
      }
    })
  }
})
