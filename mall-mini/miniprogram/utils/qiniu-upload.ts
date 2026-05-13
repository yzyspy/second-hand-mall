/**
 * 七牛云对象存储上传封装
 * 使用后端生成的 uptoken 通过 wx.uploadFile 上传文件
 * 参考文档: https://developer.qiniu.com/kodo/1312/upload
 */

import { post } from './request'

// ======================== 类型定义 ========================

interface QiniuTokenResponse {
  uploadKey: string
  upToken: string
  domain: string
  uploadUrl: string
}

export interface UploadResult {
  url: string
  cosPath: string // 保留字段名兼容性，实际是七牛云路径
}

// ======================== 上传逻辑 ========================

/**
 * 获取七牛云上传凭证
 */
async function getQiniuToken(key?: string): Promise<QiniuTokenResponse> {
  const response = await post<QiniuTokenResponse>('/api/upload/qiniu-token', key ? { key } : {})
  return response.data as QiniuTokenResponse
}

/**
 * 上传单个文件到七牛云
 * @param filePath 本地文件路径（从chooseMedia返回）
 * @param customKey 自定义上传路径（可选）
 */
export async function uploadToQiniu(filePath: string, customKey?: string): Promise<UploadResult> {
  const qiniuInfo = await getQiniuToken(customKey)

  const {
    uploadKey,
    upToken,
    domain,
    uploadUrl,
  } = qiniuInfo

  return new Promise((resolve, reject) => {
    wx.uploadFile({
      url: uploadUrl,
      filePath: filePath,
      name: 'file',
      formData: {
        key: uploadKey,
        token: upToken,
      },
      success: (res: WechatMiniprogram.UploadFileSuccessCallbackResult) => {
        if (res.statusCode === 200) {
          // 七牛云返回200表示上传成功
          resolve({
            url: `${domain}/${uploadKey}`,
            cosPath: uploadKey,
          })
        } else {
          console.error('七牛云上传失败:', res.statusCode, res.data)
          reject(new Error(`七牛云上传失败: ${res.statusCode} ${res.data}`))
        }
      },
      fail: (err) => {
        console.error('七牛云上传请求失败:', err)
        reject(new Error(`七牛云上传请求失败: ${err.errMsg}`))
      },
    })
  })
}

/**
 * 批量上传多个文件
 * @param filePaths 本地文件路径数组
 */
export async function uploadMultipleFiles(filePaths: string[]): Promise<UploadResult[]> {
  const results: UploadResult[] = []

  for (const filePath of filePaths) {
    try {
      const result = await uploadToQiniu(filePath)
      results.push(result)
    } catch (err) {
      console.error('文件上传失败:', filePath, err)
      throw err
    }
  }

  return results
}
