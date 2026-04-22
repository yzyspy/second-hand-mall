/**
 * 腾讯云COS上传封装
 * 使用STS临时密钥方式上传文件
 * 参考: https://cloud.tencent.com/document/product/436/14690
 */

import { post } from './request'

// ======================== 签名工具函数 ========================

/**
 * HMAC-SHA1 签名
 * 纯JS实现，不依赖外部库
 */
function hmacSha1(key: string | ArrayBuffer, data: string): string {
  const keyBytes = typeof key === 'string' ? strToBytes(key) : new Uint8Array(key)
  const dataBytes = strToBytes(data)
  return arrayBufferToHex(hmacSha1Raw(keyBytes, dataBytes))
}

function hmacSha1Raw(key: Uint8Array, data: Uint8Array): ArrayBuffer {
  const BLOCK_SIZE = 64

  // 如果key超过块大小，先hash
  if (key.length > BLOCK_SIZE) {
    key = new Uint8Array(sha1Raw(key))
  }

  // 补齐到块大小
  const paddedKey = new Uint8Array(BLOCK_SIZE)
  paddedKey.set(key)

  // 生成 ipad 和 opad
  const ipad = new Uint8Array(BLOCK_SIZE)
  const opad = new Uint8Array(BLOCK_SIZE)
  for (let i = 0; i < BLOCK_SIZE; i++) {
    ipad[i] = paddedKey[i] ^ 0x36
    opad[i] = paddedKey[i] ^ 0x5c
  }

  // SHA1(ipad + data)
  const innerData = new Uint8Array(BLOCK_SIZE + data.length)
  innerData.set(ipad)
  innerData.set(data, BLOCK_SIZE)
  const innerHash = new Uint8Array(sha1Raw(innerData))

  // SHA1(opad + innerHash)
  const outerData = new Uint8Array(BLOCK_SIZE + 20)
  outerData.set(opad)
  outerData.set(innerHash, BLOCK_SIZE)
  return sha1Raw(outerData)
}

/**
 * SHA1 哈希
 */
function sha1Raw(data: Uint8Array): ArrayBuffer {
  let h0 = 0x67452301
  let h1 = 0xEFCDAB89
  let h2 = 0x98BADCFE
  let h3 = 0x10325476
  let h4 = 0xC3D2E1F0

  const msgLen = data.length
  const bitLen = msgLen * 8

  // 补位 + 长度
  const newLen = Math.ceil((msgLen + 9) / 64) * 64
  const msg = new Uint8Array(newLen)
  msg.set(data)
  msg[msgLen] = 0x80

  // 写入原始长度(64位大端)
  const dv = new DataView(msg.buffer)
  dv.setUint32(newLen - 4, bitLen >>> 0, false)

  // 处理每个512位(64字节)块
  for (let offset = 0; offset < newLen; offset += 64) {
    const w = new Uint32Array(80)
    for (let i = 0; i < 16; i++) {
      w[i] = dv.getUint32(offset + i * 4, false)
    }
    for (let i = 16; i < 80; i++) {
      const x = w[i - 3] ^ w[i - 8] ^ w[i - 14] ^ w[i - 16]
      w[i] = ((x << 1) | (x >>> 31)) >>> 0
    }

    let a = h0, b = h1, c = h2, d = h3, e = h4

    for (let i = 0; i < 80; i++) {
      let f: number, k: number
      if (i < 20) {
        f = (b & c) | (~b & d)
        k = 0x5A827999
      } else if (i < 40) {
        f = b ^ c ^ d
        k = 0x6ED9EBA1
      } else if (i < 60) {
        f = (b & c) | (b & d) | (c & d)
        k = 0x8F1BBCDC
      } else {
        f = b ^ c ^ d
        k = 0xCA62C1D6
      }
      const temp = (((a << 5) | (a >>> 27)) + f + e + k + w[i]) >>> 0
      e = d
      d = c
      c = ((b << 30) | (b >>> 2)) >>> 0
      b = a
      a = temp
    }

    h0 = (h0 + a) >>> 0
    h1 = (h1 + b) >>> 0
    h2 = (h2 + c) >>> 0
    h3 = (h3 + d) >>> 0
    h4 = (h4 + e) >>> 0
  }

  const result = new ArrayBuffer(20)
  const rv = new DataView(result)
  rv.setUint32(0, h0, false)
  rv.setUint32(4, h1, false)
  rv.setUint32(8, h2, false)
  rv.setUint32(12, h3, false)
  rv.setUint32(16, h4, false)
  return result
}

function sha1Hex(data: string): string {
  return arrayBufferToHex(sha1Raw(strToBytes(data)))
}

function strToBytes(str: string): Uint8Array {
  const encoder = new (typeof TextEncoder !== 'undefined' ? TextEncoder : requirePlugin('textEncoder')) as any
  if (encoder && encoder.encode) {
    return encoder.encode(str)
  }
  // Fallback for mini-program
  const bytes: number[] = []
  for (let i = 0; i < str.length; i++) {
    const code = str.charCodeAt(i)
    if (code < 0x80) {
      bytes.push(code)
    } else if (code < 0x800) {
      bytes.push(0xc0 | (code >> 6), 0x80 | (code & 0x3f))
    } else {
      bytes.push(0xe0 | (code >> 12), 0x80 | ((code >> 6) & 0x3f), 0x80 | (code & 0x3f))
    }
  }
  return new Uint8Array(bytes)
}

function arrayBufferToHex(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer)
  let hex = ''
  for (let i = 0; i < bytes.length; i++) {
    hex += bytes[i].toString(16).padStart(2, '0')
  }
  return hex
}

// ======================== COS 签名计算 ========================

/**
 * 计算COS Authorization请求头
 * 文档: https://cloud.tencent.com/document/product/436/7778
 */
function getAuthorization(
  secretId: string,
  secretKey: string,
  sessionToken: string,
  method: string,
  path: string,
  host: string,
  startTime: number,
  expireSeconds: number
): Record<string, string> {
  const expiredTime = startTime + expireSeconds
  const keyTime = `${startTime};${expiredTime}`

  // 1. SignKey = HMAC-SHA1(SecretKey, KeyTime)
  const signKey = hmacSha1(secretKey, keyTime)

  // 2. HttpString
  const httpMethod = method.toLowerCase()
  const uriPathname = path.startsWith('/') ? path : '/' + path
  const httpParameters = ''
  const httpHeaders = `host=${host.toLowerCase()};x-cos-security-token=${sessionToken.toLowerCase()}`
  const httpString = `${httpMethod}\n${uriPathname}\n${httpParameters}\n${httpHeaders}\n`

  // 3. StringToSign
  const stringToSign = `sha1\n${keyTime}\n${sha1Hex(httpString)}\n`

  // 4. Signature
  const signature = hmacSha1(signKey, stringToSign)

  // 5. Authorization
  const headerList = 'host;x-cos-security-token'
  const parameterList = ''
  const authorization = [
    `q-sign-algorithm=sha1`,
    `q-ak=${secretId}`,
    `q-sign-time=${keyTime}`,
    `q-key-time=${keyTime}`,
    `q-header-list=${headerList}`,
    `q-url-param-list=${parameterList}`,
    `q-signature=${signature}`
  ].join('&')

  return {
    'Authorization': authorization,
    'x-cos-security-token': sessionToken
  }
}

// ======================== 上传逻辑 ========================

interface CosSignatureResponse {
  cosHost: string
  tmpSecretId: string
  tmpSecretKey: string
  sessionToken: string
  region: string
  bucket: string
  cosPath: string
  cosPathPrefix: string
}

export interface UploadResult {
  url: string
  cosPath: string
}

/**
 * 获取COS上传签名(STS临时密钥)
 */
async function getCosSignature(key?: string): Promise<CosSignatureResponse> {
  const response = await post<CosSignatureResponse>('/api/upload/cos-signature-v2', key ? { key } : {})
  return response.data as CosSignatureResponse
}

/**
 * 上传单个文件到COS (PUT Object方式)
 * @param filePath 本地文件路径（从chooseMedia返回）
 * @param customKey 自定义上传路径（可选）
 */
export async function uploadToCos(filePath: string, customKey?: string): Promise<UploadResult> {
  // 1. 获取STS临时密钥
  const cosInfo = await getCosSignature(customKey)

  const {
    cosHost,
    tmpSecretId,
    tmpSecretKey,
    sessionToken,
    bucket,
    region,
    cosPath
  } = cosInfo

  // 2. 读取文件数据
  const fs = wx.getFileSystemManager()
  const fileData = fs.readFileSync(filePath)

  // 3. 计算Authorization签名
  const host = `${bucket}.cos.${region}.myqcloud.com`
  const now = Math.floor(Date.now() / 1000)
  const headers = getAuthorization(
    tmpSecretId,
    tmpSecretKey,
    sessionToken,
    'PUT',
    '/' + cosPath,
    host,
    now,
    600
  )

  // 4. 使用wx.request PUT上传
  return new Promise((resolve, reject) => {
    wx.request({
      url: `https://${host}/${cosPath}`,
      method: 'PUT',
      data: fileData,
      header: {
        ...headers,
        'Content-Type': 'image/jpeg',
        'Host': host
      },
      success: (res: WechatMiniprogram.RequestSuccessCallbackResult) => {
        if (res.statusCode === 200) {
          resolve({
            url: `${cosHost}/${cosPath}`,
            cosPath: cosPath
          })
        } else {
          reject(new Error(`COS上传失败: ${res.statusCode}`))
        }
      },
      fail: (err) => {
        reject(new Error(`COS上传请求失败: ${err.errMsg}`))
      }
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
      const result = await uploadToCos(filePath)
      results.push(result)
    } catch (err) {
      console.error('文件上传失败:', filePath, err)
    }
  }

  return results
}