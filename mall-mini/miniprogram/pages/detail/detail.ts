// pages/detail/detail.ts

import { get } from '../../utils/request'

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
  favorites: number
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
      favorites: 0
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

        // 将逗号分隔的图片字符串转为数组
        const images = data.images ? data.images.split(',').filter((img: string) => img) : []

        const product: Product = {
          id: data.id,
          title: data.title,
          description: data.description || '',
          price: data.price,
          images: images,
          category: '二手好物',
          condition: '九成新',
          location: data.location || '',
          seller: {
            id: String(data.id),
            name: data.seller || '微信用户',
            avatar: data.avatar || '',
            rating: 4.8
          },
          createdAt: data.create_time ? data.create_time.substring(0, 10) : '',
          views: Math.floor(Math.random() * 500) + 50, // 暂时模拟
          favorites: Math.floor(Math.random() * 50) + 1 // 暂时模拟
        }

        this.setData({
          product,
          loading: false
        })
      } else {
        this.setData({
          error: response.msg || '加载失败',
          loading: false
        })
      }
    } catch (err) {
      console.error('加载商品详情失败:', err)
      this.setData({
        error: '网络错误，请重试',
        loading: false
      })
    }

    wx.stopPullDownRefresh()
  },

  toggleFavorite() {
    const isFavorite = this.data.isFavorite
    this.setData({ isFavorite: !isFavorite })
    wx.showToast({ title: isFavorite ? '已移除收藏' : '已收藏', icon: 'success' })
  },

  contactSeller() {
    wx.showToast({ title: '功能开发中', icon: 'none' })
  },

  reportProduct() {
    wx.showActionSheet({
      itemList: ['虚假信息', '骚扰信息', '违法违规'],
      success: (res) => {
        wx.showToast({ title: '举报成功', icon: 'success' })
      }
    })
  }
})
