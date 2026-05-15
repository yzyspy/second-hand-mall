// pages/productEdit/productEdit.ts
import { get, put } from '../../utils/request'
import { uploadToQiniu } from '../../utils/qiniu-upload'
import regionsData from '../../data/china-regions'

interface UploadedImage {
  localPath: string
  remoteUrl?: string
  uploading?: boolean
}

interface EditData {
  productId: number
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
}

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

// 将 location 字符串解析回省/市/区三级索引
function parseLocationToIndexes(locationStr: string): number[] {
  for (let pi = 0; pi < regionsData.length; pi++) {
    const province = regionsData[pi]
    if (!locationStr.startsWith(province.name)) continue
    const afterProvince = locationStr.slice(province.name.length)
    const cities = province.children
    for (let ci = 0; ci < cities.length; ci++) {
      const city = cities[ci]
      // 直辖市：province.name === city.name，location = province + district
      if (province.name === city.name) {
        const districts = city.children as string[]
        for (let di = 0; di < districts.length; di++) {
          if (afterProvince === districts[di]) return [pi, ci, di]
        }
        return [pi, ci, 0]
      }
      if (!afterProvince.startsWith(city.name)) continue
      const afterCity = afterProvince.slice(city.name.length)
      const districts = city.children as string[]
      for (let di = 0; di < districts.length; di++) {
        if (afterCity === districts[di]) return [pi, ci, di]
      }
      return [pi, ci, 0]
    }
  }
  return [0, 0, 0]
}

const CATEGORIES = ['电子产品', '服装鞋帽', '图书文具', '生活用品', '数码配件', '其他']

Page<EditData, WechatMiniprogram.IAnyObject>({
  data: {
    productId: 0,
    images: [],
    maxImages: 9,
    description: '',
    price: '',
    location: '',
    regionNames: buildInitialRegionNames(),
    regionIndexes: [0, 0, 0],
    categoryIndex: 0,
    categories: CATEGORIES,
    submitting: false
  },

  async onLoad(options: Record<string, string>) {
    const id = Number(options.id)
    if (!id) {
      wx.showToast({ title: '商品不存在', icon: 'none' })
      return
    }
    this.setData({ productId: id })
    await this.loadProduct(id)
  },

  async loadProduct(id: number) {
    wx.showLoading({ title: '加载中...', mask: true })
    try {
      const res = await get(`/api/product/detail?id=${id}`)
      const p = res.data
      // 图片回填：已上传的 remote URL 用 localPath 展示，remoteUrl 标记已上传
      const images: UploadedImage[] = p.images
        ? p.images.split(',').filter(Boolean).map((url: string) => ({
            localPath: url,
            remoteUrl: url
          }))
        : []

      // location 解析回索引
      const [pi, ci, di] = parseLocationToIndexes(p.location || '')
      const regionNames = [provinceNames, getCityNames(pi), getDistrictNames(pi, ci)]

      // 分类匹配
      const categoryIndex = Math.max(0, CATEGORIES.indexOf(p.category || ''))

      this.setData({
        images,
        description: p.description || '',
        price: p.price ? String(p.price) : '',
        location: p.location || '',
        regionNames,
        regionIndexes: [pi, ci, di],
        categoryIndex
      })
    } catch (err) {
      wx.showToast({ title: '加载商品失败', icon: 'none' })
    } finally {
      wx.hideLoading()
    }
  },

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
      const newImages: UploadedImage[] = res.tempFiles.map(f => ({
        localPath: f.tempFilePath,
        uploading: false
      }))
      this.setData({ images: [...images, ...newImages] })
    } catch (_) {}
  },

  deleteImage(e: WechatMiniprogram.TouchEvent) {
    const { index } = e.currentTarget.dataset
    const images = [...this.data.images]
    images.splice(index, 1)
    this.setData({ images })
  },

  previewImage(e: WechatMiniprogram.TouchEvent) {
    const { index } = e.currentTarget.dataset
    const urls = this.data.images.map(img => img.localPath)
    wx.previewImage({ current: urls[index], urls })
  },

  onDescriptionInput(e: WechatMiniprogram.InputEvent) {
    this.setData({ description: e.detail.value })
  },

  onPriceInput(e: WechatMiniprogram.InputEvent) {
    const formatted = e.detail.value
      .replace(/[^\d.]/g, '')
      .replace(/\.{2,}/g, '.')
      .replace(/^(\d+\.\d{2}).*$/, '$1')
    this.setData({ price: formatted })
  },

  onCategoryChange(e: WechatMiniprogram.PickerChange) {
    this.setData({ categoryIndex: Number(e.detail.value) })
  },

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

  onRegionChange(e: WechatMiniprogram.PickerChange) {
    const [pi, ci, di] = e.detail.value as number[]
    const province = regionsData[pi].name
    const city = regionsData[pi].children[ci].name
    const district = (regionsData[pi].children[ci].children as string[])[di]
    const location = province === city ? `${province}${district}` : `${province}${city}${district}`
    this.setData({ regionIndexes: [pi, ci, di], location })
  },

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
    return true
  },

  async uploadImages(): Promise<string[]> {
    const { images } = this.data
    const urls: string[] = []
    for (let i = 0; i < images.length; i++) {
      const img = images[i]
      if (img.remoteUrl) {
        urls.push(img.remoteUrl)
        continue
      }
      this.setData({ [`images[${i}].uploading`]: true })
      try {
        const result = await uploadToQiniu(img.localPath)
        urls.push(result.url)
        this.setData({
          [`images[${i}].remoteUrl`]: result.url,
          [`images[${i}].uploading`]: false
        })
      } catch (err) {
        this.setData({ [`images[${i}].uploading`]: false })
        throw err
      }
    }
    return urls
  },

  async submitForm() {
    if (!this.validateForm()) return
    if (this.data.submitting) return
    this.setData({ submitting: true })
    wx.showLoading({ title: '保存中...', mask: true })
    try {
      const imageUrls = await this.uploadImages()
      await put('/api/product/update', {
        id: this.data.productId,
        description: this.data.description,
        price: parseFloat(this.data.price),
        location: this.data.location,
        images: imageUrls
      })
      wx.hideLoading()
      wx.showToast({ title: '保存成功', icon: 'success' })
      setTimeout(() => wx.navigateBack(), 1500)
    } catch (err) {
      wx.hideLoading()
      wx.showToast({ title: '保存失败，请重试', icon: 'none' })
      console.error('保存失败:', err)
    } finally {
      this.setData({ submitting: false })
    }
  }
})
