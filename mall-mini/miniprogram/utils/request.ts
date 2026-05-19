/**
 * HTTP请求封装
 * 封装wx.request，支持Promise，自动注入JWT token
 */
// 修改ip地址
//export const BASE_URL = 'http://localhost:8080'
//export const BASE_URL = "http://101.34.238.222:80"
export const BASE_URL = "https://yangzhongyu.site"
//export const IMG_BASE_URL = "https://wwww.yangzhongyu.site"

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

/**
 * 静默微信重登录，获取新 token 并写入 storage。
 * 失败时清除 storage 中的认证信息并抛出错误。
 * 直接使用 wx.request 而非 request() 封装，避免 token 注入和错误 toast 干扰重登录流程。
 */
export async function silentReLogin(): Promise<void> {
  const clearAuth = () => {
    wx.removeStorageSync('token')
    wx.removeStorageSync('userInfo')
    wx.removeStorageSync('userId')
  }

  const loginRes = await new Promise<WechatMiniprogram.LoginSuccessCallbackResult>(
    (resolve, reject) => wx.login({ success: resolve, fail: reject })
  )

  if (!loginRes.code) {
    clearAuth()
    throw new Error('获取登录凭证失败')
  }

  const res = await new Promise<WechatMiniprogram.RequestSuccessCallbackResult>(
    (resolve, reject) =>
      wx.request({
        url: BASE_URL + '/api/user/wx-login',
        method: 'POST',
        data: { code: loginRes.code },
        header: { 'Content-Type': 'application/json' },
        success: resolve,
        fail: reject,
      })
  )

  const response = res.data as ApiResponse<{ token: string; avatar: string; nick_name: string; user_id: number }>

  if (response.code !== 0) {
    clearAuth()
    throw new Error(response.msg || '登录失败')
  }

  if (!response.data) {
    throw new Error('登录响应数据缺失')
  }

  wx.setStorageSync('token', response.data.token)
  wx.setStorageSync('userInfo', {
    avatarUrl: response.data.avatar || '/images/default-avatar.png',
    nickName: response.data.nick_name || '微信用户',
  })
  wx.setStorageSync('userId', response.data.user_id)
}

/**
 * 内部实现，支持重试标志避免 401 死循环
 */
async function _request<T = any>(options: RequestOptions, isRetry: boolean): Promise<ApiResponse<T>> {
  const { url, method = 'GET', data, header = {}, showLoading = false, showError = true } = options

  if (showLoading) {
    wx.showLoading({ title: '加载中...', mask: true })
  }

  const token = wx.getStorageSync('token')
  if (token) {
    header['Authorization'] = `Bearer ${token}`
  }

  if (!header['Content-Type'] && method !== 'GET') {
    header['Content-Type'] = 'application/json'
  }

  return new Promise((resolve, reject) => {
    wx.request({
      url: BASE_URL + url,
      method,
      data,
      header,
      success: async (res) => {
        if (showLoading) {
          wx.hideLoading()
        }

        // 401: token 过期，尝试静默重登录（只重试一次）
        if (res.statusCode === 401 && !isRetry) {
          try {
            await silentReLogin()
            resolve(await _request<T>(options, true))
          } catch {
            wx.showToast({ title: '登录已过期，请重新登录', icon: 'none' })
            reject(new Error('登录已过期'))
          }
          return
        }

        const response = res.data as ApiResponse<T>

        if (response.code === 0) {
          resolve(response)
          return
        }

        if (showError) {
          wx.showToast({ title: response.msg || '请求失败', icon: 'none' })
        }
        reject(response)
      },
      fail: (err) => {
        if (showLoading) {
          wx.hideLoading()
        }
        if (showError) {
          wx.showToast({ title: '网络请求失败', icon: 'none' })
        }
        reject(err)
      },
    })
  })
}

/**
 * 发起HTTP请求
 */
export function request<T = any>(options: RequestOptions): Promise<ApiResponse<T>> {
  return _request<T>(options, false)
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

/**
 * PUT请求
 */
export function put<T = any>(url: string, data?: any): Promise<ApiResponse<T>> {
  return request<T>({ url, method: 'PUT', data })
}