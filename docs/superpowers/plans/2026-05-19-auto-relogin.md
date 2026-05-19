# Auto Silent Re-Login Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** When any API call returns HTTP 401, silently re-login via WeChat and retry the request once — with zero changes to page-level code.

**Architecture:** Split `request()` into a public wrapper and an internal `_request(options, isRetry)`. Export a new `silentReLogin()` function that handles wx.login + backend call + storage. The 401 intercept lives in `_request()`; `my.ts` reuses `silentReLogin()` instead of duplicating the login flow.

**Tech Stack:** TypeScript, WeChat Mini Program APIs (`wx.login`, `wx.request`, `wx.setStorageSync`), existing `request.ts` utility.

---

### Task 1: Add `silentReLogin()` to `request.ts`

**Files:**
- Modify: `mall-mini/miniprogram/utils/request.ts`

- [ ] **Step 1: Add `silentReLogin()` at the bottom of `request.ts`**

Open `mall-mini/miniprogram/utils/request.ts`. Append the following export after the existing `put()` function:

```typescript
/**
 * 静默微信重登录，获取新 token 并写入 storage。
 * 失败时清除 storage 中的认证信息并抛出错误。
 */
export async function silentReLogin(): Promise<void> {
  const loginRes = await new Promise<WechatMiniprogram.LoginSuccessCallbackResult>(
    (resolve, reject) => wx.login({ success: resolve, fail: reject })
  )

  if (!loginRes.code) {
    wx.removeStorageSync('token')
    wx.removeStorageSync('userInfo')
    wx.removeStorageSync('userId')
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

  const response = res.data as ApiResponse<{
    token: string
    avatar: string
    nick_name: string
    user_id: number
  }>

  if (response.code !== 0) {
    wx.removeStorageSync('token')
    wx.removeStorageSync('userInfo')
    wx.removeStorageSync('userId')
    throw new Error(response.msg || '登录失败')
  }

  wx.setStorageSync('token', response.data!.token)
  wx.setStorageSync('userInfo', {
    avatarUrl: response.data!.avatar || '/images/default-avatar.png',
    nickName: response.data!.nick_name || '微信用户',
  })
  wx.setStorageSync('userId', response.data!.user_id)
}
```

- [ ] **Step 2: Verify TypeScript compiles without errors**

In WeChat Developer Tools, open the project and check the compiler output panel at the bottom. There should be no red TypeScript errors in `utils/request.ts`.

- [ ] **Step 3: Commit**

```bash
git add mall-mini/miniprogram/utils/request.ts
git commit -m "feat: add silentReLogin() to request.ts"
```

---

### Task 2: Add `_request()` internal function with 401 intercept

**Files:**
- Modify: `mall-mini/miniprogram/utils/request.ts`

- [ ] **Step 1: Replace the current `request()` export with `_request()` + thin wrapper**

Find the current `export function request<T = any>(options: RequestOptions)` in `request.ts` and replace the entire function with the following two functions (keep everything else in the file untouched):

```typescript
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
```

- [ ] **Step 2: Check function order in the file**

The final order of functions in `request.ts` should be:

1. `BASE_URL` constant + interfaces
2. `silentReLogin()` — exported, readable order (hoisting means any order works)
3. `_request()` — internal, calls `silentReLogin`
4. `request()` — public wrapper
5. `get()`, `post()`, `put()` — unchanged

- [ ] **Step 3: Verify TypeScript compiles without errors**

In WeChat Developer Tools, check the compiler output. No red errors in `utils/request.ts`.

- [ ] **Step 4: Commit**

```bash
git add mall-mini/miniprogram/utils/request.ts
git commit -m "feat: intercept 401 in _request() and auto retry after silentReLogin"
```

---

### Task 3: Refactor `my.ts` `handleLogin()` to use `silentReLogin()`

**Files:**
- Modify: `mall-mini/miniprogram/pages/my/my.ts`

- [ ] **Step 1: Update the import line in `my.ts`**

Find the current import at the top of `my.ts`:

```typescript
import { post } from '../../utils/request'
import {BASE_URL} from '../../utils/request'
```

Replace with a single import that includes `silentReLogin`:

```typescript
import { silentReLogin, BASE_URL } from '../../utils/request'
```

- [ ] **Step 2: Replace `handleLogin()` body**

Find the current `handleLogin()` method and replace its entire body:

```typescript
async handleLogin() {
  wx.showLoading({ title: '登录中...', mask: true })
  try {
    await silentReLogin()
    const userInfo = wx.getStorageSync('userInfo')
    this.setData({ userInfo, isLoggedIn: true })
    wx.showToast({ title: '登录成功', icon: 'success' })
  } catch (err: any) {
    console.log('登录失败', err)
    console.log('' + BASE_URL)
    wx.showToast({ title: err.message || '登录失败', icon: 'none' })
  } finally {
    wx.hideLoading()
  }
},
```

- [ ] **Step 3: Verify TypeScript compiles without errors**

In WeChat Developer Tools, check the compiler output. No red errors in `pages/my/my.ts`.

- [ ] **Step 4: Commit**

```bash
git add mall-mini/miniprogram/pages/my/my.ts
git commit -m "refactor: my.ts handleLogin() delegates to silentReLogin()"
```

---

### Task 4: Manual Verification

**No code changes — verification only.**

- [ ] **Step 1: Simulate expired token**

In WeChat Developer Tools → Storage panel, find the `token` key and replace its value with any invalid/expired JWT string, e.g.:

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJleHAiOjE2MDAwMDAwMDB9.invalid
```

- [ ] **Step 2: Trigger an authenticated request**

Navigate to any page that calls an auth-required API (e.g., "我发布的" → calls `GET /api/product/mine`, or "我的收藏" → calls `GET /api/favorite/list`).

- [ ] **Step 3: Verify silent re-login succeeded**

Expected behavior:
- No "登录已过期" toast appears
- The page data loads normally
- In the Storage panel, `token` now contains a new JWT value (different from the expired one you set)

- [ ] **Step 4: Verify the "我的" page login flow still works**

Log out (remove token from Storage), go to "我的" tab, tap the login button. Verify login succeeds and user info displays correctly.

- [ ] **Step 5: Verify failure path**

In WeChat Developer Tools → Network panel, enable "Offline" mode or block requests to `yangzhongyu.site`. Set an expired token in Storage, then trigger an authenticated request.

Expected behavior:
- Toast "登录已过期，请重新登录" appears
- `token`, `userInfo`, `userId` are cleared from Storage
