// pages/publish/publish.ts

import { uploadToCos } from '../../utils/cos-upload'

interface UploadedImage {
  localPath: string
  remoteUrl?: string
  uploading?: boolean
  uploadError?: string
}

interface PublishData {
  images: UploadedImage[]
  maxImages: number
  description: string
  price: string
  location: string
  categoryIndex: number
  categories: string[]
  submitting: boolean
}

Page<PublishData, WechatMiniprogram.IAnyObject>({
  data: {
    images: [],
    maxImages: 9,
    description: '',
    price: '',
    location: '',
    categoryIndex: 0,
    categories: ['电子产品', '服装鞋帽', '图书文具', '生活用品', '数码配件', '其他'],
    submitting: false
  },

  onLoad() {
    // 请求用户位置权限（可选）
  },

  onShow() {
    // 页面显示
  },

  // 选择图片
  async chooseImage() {
    const { images, maxImages } = this.data
    const remaining = maxImages - images.length

    if (remaining <= 0) {
      wx.showToast({ title: `最多上传${maxImages}张图片`, icon: 'none' })
      return
    }

    try {
      const res = await wx.chooseMedia({
        count: remaining,
        mediaType: ['image'],
        sourceType: ['album', 'camera'],
        sizeType: ['compressed']
      })

      const newImages: UploadedImage[] = res.tempFiles.map(file => ({
        localPath: file.tempFilePath,
        uploading: false
      }))

      this.setData({
        images: [...images, ...newImages]
      })
    } catch (err) {
      console.log('取消选择图片', err)
    }
  },

  // 删除图片
  deleteImage(e: WechatMiniprogram.TouchEvent) {
    const { index } = e.currentTarget.dataset
    const images = this.data.images
    images.splice(index, 1)
    this.setData({ images })
  },

  // 预览图片
  previewImage(e: WechatMiniprogram.TouchEvent) {
    const { index } = e.currentTarget.dataset
    const urls = this.data.images.map(img => img.localPath)
    wx.previewImage({
      current: urls[index],
      urls: urls
    })
  },

  // 输入描述
  onDescriptionInput(e: WechatMiniprogram.InputEvent) {
    this.setData({ description: e.detail.value })
  },

  // 输入价格
  onPriceInput(e: WechatMiniprogram.InputEvent) {
    const value = e.detail.value
    // 只允许数字和小数点，最多两位小数
    const formatted = value.replace(/[^\d.]/g, '')
      .replace(/\.{2,}/g, '.')
      .replace(/^(\d+\.\d{2}).*$/, '$1')
    this.setData({ price: formatted })
  },

  // 输入地点
  onLocationInput(e: WechatMiniprogram.InputEvent) {
    this.setData({ location: e.detail.value })
  },

  // 选择分类
  onCategoryChange(e: WechatMiniprogram.PickerChange) {
    this.setData({ categoryIndex: Number(e.detail.value) })
  },

  // 获取当前位置
  async getLocation() {
    try {
      const res = await wx.getLocation({ type: 'gcj02' })
      // 这里可以调用逆地理编码API获取地址名称
      // 暂时显示坐标
      this.setData({ location: `${res.latitude.toFixed(4)}, ${res.longitude.toFixed(4)}` })
    } catch (err) {
      wx.showToast({ title: '获取位置失败', icon: 'none' })
    }
  },

  // 验证表单
  validateForm(): boolean {
    const { images, description, price, location } = this.data

    if (images.length === 0) {
      wx.showToast({ title: '请至少上传一张图片', icon: 'none' })
      return false
    }

    if (!description.trim()) {
      wx.showToast({ title: '请填写商品描述', icon: 'none' })
      return false
    }

    if (!price) {
      wx.showToast({ title: '请填写价格', icon: 'none' })
      return false
    }

    if (!location.trim()) {
      wx.showToast({ title: '请填写交易地点', icon: 'none' })
      return false
    }

    return true
  },

  // 上传所有图片
  async uploadImages(): Promise<string[]> {
    const { images } = this.data
    const uploadedUrls: string[] = []

    for (let i = 0; i < images.length; i++) {
      const img = images[i]

      // 如果已经上传过，跳过
      if (img.remoteUrl) {
        uploadedUrls.push(img.remoteUrl)
        continue
      }

      // 标记正在上传
      this.setData({ [`images[${i}].uploading`]: true })

      try {
        const result = await uploadToCos(img.localPath)
        uploadedUrls.push(result.url)
        this.setData({
          [`images[${i}].remoteUrl`]: result.url,
          [`images[${i}].uploading`]: false
        })
      } catch (err) {
        this.setData({
          [`images[${i}].uploading`]: false,
          [`images[${i}].uploadError`]: '上传失败'
        })
        throw err
      }
    }

    return uploadedUrls
  },

  // 提交表单
  async submitForm() {
    if (!this.validateForm()) return
    if (this.data.submitting) return

    this.setData({ submitting: true })

    wx.showLoading({ title: '发布中...', mask: true })

    try {
      // 1. 上传所有图片
      const imageUrls = await this.uploadImages()

      // 2. 构建商品数据
      const productData = {
        title: this.data.description.substring(0, 50), // 取描述前50字作为标题
        description: this.data.description,
        price: parseFloat(this.data.price),
        location: this.data.location,
        category: this.data.categories[this.data.categoryIndex],
        images: imageUrls
      }

      console.log('提交商品数据:', productData)

      // 3. 调用后端API保存商品 (需要后端实现对应接口)
      // const result = await post('/api/product/publish', productData)

      // 模拟成功
      await new Promise(resolve => setTimeout(resolve, 1000))

      wx.hideLoading()
      wx.showToast({ title: '发布成功', icon: 'success' })

      // 清空表单
      setTimeout(() => {
        this.setData({
          images: [],
          description: '',
          price: '',
          location: '',
          categoryIndex: 0
        })
      }, 1500)

    } catch (err) {
      wx.hideLoading()
      wx.showToast({ title: '发布失败，请重试', icon: 'none' })
      console.error('发布失败:', err)
    } finally {
      this.setData({ submitting: false })
    }
  }
})
