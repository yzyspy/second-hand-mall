// pages/publish/publish.ts

import { post } from '../../utils/request'
import { uploadToQiniu } from '../../utils/qiniu-upload'
import regionsData from '../../data/china-regions'

const provinceNames = regionsData.map(p => p.name)

function getCityNames(pi: number): string[] {
  return regionsData[pi].children.map(c => c.name)
}

function getDistrictNames(pi: number, ci: number): string[] {
  return regionsData[pi].children[ci].children as string[]
}

function buildInitialRegionNames(): string[][] {
  return [provinceNames, getCityNames(0), getDistrictNames(0, 0)]
}

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
  regionNames: string[][]
  regionIndexes: number[]
  categoryIndex: number
  categories: string[]
  submitting: boolean
  contactType: 'phone' | 'wechat' | 'qq'
  contactValue: string
  contactTypes: string[]
}

Page<PublishData, WechatMiniprogram.IAnyObject>({
  data: {
    images: [],
    maxImages: 9,
    description: '',
    price: '',
    location: '',
    regionNames: buildInitialRegionNames(),
    regionIndexes: [0, 0, 0],
    categoryIndex: 0,
    categories: ['电子产品', '服装鞋帽', '图书文具', '生活用品', '数码配件', '其他'],
    submitting: false,
    contactType: 'phone',
    contactValue: '',
    contactTypes: ['手机号', '微信', 'QQ'],
  },

  onLoad() {
    // 请求用户位置权限（可选）
  },

  onShow() {
    // 检查登录状态
    const token = wx.getStorageSync('token')
    if (!token) {
      wx.showModal({
        title: '请先登录',
        content: '发布商品需要登录',
        confirmText: '去登录',
        success: (res) => {
          if (res.confirm) {
            wx.switchTab({ url: '/pages/my/my' })
          }
        }
      })
    }
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

  // 级联选择器列变化（滚动时实时更新联动列）
  onRegionColumnChange(e: WechatMiniprogram.PickerColumnChange) {
    const { column, value } = e.detail
    const indexes = [...this.data.regionIndexes]
    indexes[column] = value
    if (column === 0) {
      indexes[1] = 0
      indexes[2] = 0
      this.setData({
        regionIndexes: indexes,
        'regionNames[1]': getCityNames(value),
        'regionNames[2]': getDistrictNames(value, 0)
      })
    } else if (column === 1) {
      indexes[2] = 0
      this.setData({
        regionIndexes: indexes,
        'regionNames[2]': getDistrictNames(indexes[0], value)
      })
    } else {
      this.setData({ regionIndexes: indexes })
    }
  },

  // 确认选择地区
  onRegionChange(e: WechatMiniprogram.PickerChange) {
    const [pi, ci, di] = e.detail.value as number[]
    const province = regionsData[pi].name
    const city = regionsData[pi].children[ci].name
    const district = (regionsData[pi].children[ci].children as string[])[di]
    const location = province === city ? `${province}${district}` : `${province}${city}${district}`
    this.setData({ regionIndexes: [pi, ci, di], location })
  },

  // 选择分类
  onCategoryChange(e: WechatMiniprogram.PickerChange) {
    this.setData({ categoryIndex: Number(e.detail.value) })
  },

  onContactTypeSelect(e: WechatMiniprogram.TouchEvent) {
    const types: Array<'phone' | 'wechat' | 'qq'> = ['phone', 'wechat', 'qq']
    this.setData({ contactType: types[e.currentTarget.dataset.index] })
  },

  onContactValueInput(e: WechatMiniprogram.Input) {
    this.setData({ contactValue: e.detail.value })
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
      wx.showToast({ title: '请选择交易地点', icon: 'none' })
      return false
    }

    if (!this.data.contactValue.trim()) {
      wx.showToast({ title: '请填写联系方式', icon: 'none' })
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
        const result = await uploadToQiniu(img.localPath)
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
        console.log("upload image error", err)
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
        images: imageUrls,
        contact_type: this.data.contactType,
        contact_value: this.data.contactValue,
      }

      console.log('提交商品数据:', productData)

      // 3. 调用后端API保存商品
      await post('/api/product/publish', productData)

      wx.hideLoading()
      wx.showToast({ title: '发布成功', icon: 'success' })

      // 清空表单
      setTimeout(() => {
        this.setData({
          images: [],
          description: '',
          price: '',
          location: '',
          regionNames: buildInitialRegionNames(),
          regionIndexes: [0, 0, 0],
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
