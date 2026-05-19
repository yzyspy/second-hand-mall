# Auto Silent Re-Login on Token Expiry

**Date:** 2026-05-19
**Status:** Approved

## Problem

JWT token expires after 24 hours. WeChat session remains valid, but the user is forced to manually log out and log back in. All API calls requiring auth silently fail or show error toasts without explaining the cause.

## Goal

When a 401 is returned due to token expiry, automatically and silently re-login via WeChat and retry the original request — with zero changes required in any page-level code.

## Approach

Intercept HTTP 401 in `request.ts`. On 401, call `silentReLogin()` to get a new token via `wx.login()` + `/api/user/wx-login`, then retry the original request once. If re-login fails, clear the stale token and show a Toast.

## Architecture

### Files Changed

| File | Change |
|------|--------|
| `mall-mini/miniprogram/utils/request.ts` | Add `silentReLogin()`, add `isRetry` flag, intercept 401 in success callback |
| `mall-mini/miniprogram/pages/my/my.ts` | Replace inline login logic in `handleLogin()` with a call to `silentReLogin()` |

### Data Flow

```
request(options)
  → inject token from storage
  → wx.request(...)
  → success callback:
      if res.statusCode === 401 AND !isRetry:
          silentReLogin()
              wx.login() → code
              POST /api/user/wx-login { code }
              → success: setStorageSync(token, ...), retry request(options, isRetry=true)
              → failure: removeStorageSync(token/userInfo/userId)
                         showToast("登录已过期，请重新登录")
                         reject(err)
      else:
          original code===0 logic (unchanged)
```

### `silentReLogin()` Signature

```typescript
async function silentReLogin(): Promise<void>
```

- Calls `wx.login()` to get a fresh WeChat code
- POSTs to `/api/user/wx-login` with `{ code }`
- On success: saves `token`, `userInfo`, `userId` to storage
- On failure: clears those storage keys, throws error

### Retry Guard

`request()` accepts an internal `isRetry: boolean` parameter (not exposed in `RequestOptions`). The 401 intercept only triggers when `isRetry` is `false`, so a second 401 after re-login goes straight to the failure path — no infinite loop.

### `my.ts` Refactor

`handleLogin()` in `my.ts` calls `silentReLogin()` instead of duplicating the wx.login flow. This ensures a single source of truth for the login sequence.

## Error Handling

| Scenario | Behavior |
|----------|----------|
| Token expired, re-login succeeds | Retry original request transparently, user sees no interruption |
| Token expired, `wx.login()` fails | Clear token, Toast "登录已过期，请重新登录" |
| Token expired, `/api/user/wx-login` fails | Clear token, Toast "登录已过期，请重新登录" |
| Retry request gets 401 again | No further retry, Toast "登录已过期，请重新登录" |
| Non-401 errors | Existing behavior unchanged |

## Testing

1. Replace the stored token with an expired JWT string in WeChat DevTools storage panel
2. Navigate to any page that makes an authenticated API call
3. Verify: request succeeds silently, no toast, data loads normally
4. Verify: new token is stored in storage after auto re-login
5. Simulate re-login failure (e.g., disconnect network before wx.login resolves): verify Toast appears and token is cleared
