/**
 * HTTP请求封装
 * 封装wx.request，支持Promise，自动注入JWT token
 */

interface RequestOptions {
  url: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: any
  header?: Record<string, string>
  showLoading?: boolean
  showError?: boolean
}

interface ApiResponse<T = any> {
  code: number
  msg: string
  data?: T
}

const BASE_URL = 'http://localhost:8080'

/**
 * 发起HTTP请求
 */
export function request<T = any>(options: RequestOptions): Promise<ApiResponse<T>> {
  const { url, method = 'GET', data, header = {}, showLoading = false, showError = true } = options

  if (showLoading) {
    wx.showLoading({ title: '加载中...', mask: true })
  }

  // 自动注入JWT token
  const token = wx.getStorageSync('token')
  if (token) {
    header['Authorization'] = `Bearer ${token}`
  }

  // 设置Content-Type
  if (!header['Content-Type'] && method !== 'GET') {
    header['Content-Type'] = 'application/json'
  }

  return new Promise((resolve, reject) => {
    wx.request({
      url: BASE_URL + url,
      method,
      data,
      header,
      success: (res) => {
        if (showLoading) {
          wx.hideLoading()
        }

        const response = res.data as ApiResponse<T>

        // 业务成功
        if (response.code === 0) {
          resolve(response)
          return
        }

        // 业务失败
        if (showError) {
          wx.showToast({
            title: response.msg || '请求失败',
            icon: 'none'
          })
        }
        reject(response)
      },
      fail: (err) => {
        if (showLoading) {
          wx.hideLoading()
        }

        if (showError) {
          wx.showToast({
            title: '网络请求失败',
            icon: 'none'
          })
        }
        reject(err)
      }
    })
  })
}

/**
 * GET请求
 */
export function get<T = any>(url: string, data?: any): Promise<ApiResponse<T>> {
  return request<T>({ url, method: 'GET', data })
}

/**
 * POST请求
 */
export function post<T = any>(url: string, data?: any): Promise<ApiResponse<T>> {
  return request<T>({ url, method: 'POST', data })
}