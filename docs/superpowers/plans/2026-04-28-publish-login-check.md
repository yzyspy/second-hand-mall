# 发布商品前登录检查实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在用户进入发布页面时检查登录状态，未登录时弹窗引导去个人中心登录

**Architecture:** 在发布页面的 onShow 生命周期中检查 token，未登录时显示模态弹窗，点击确定跳转个人中心

**Tech Stack:** 微信小程序 TypeScript, wx.showModal, wx.switchTab

---

### Task 1: 添加登录检查逻辑

**Files:**
- Modify: `mall-mini/miniprogram/pages/publish/publish.ts`

- [ ] **Step 1: 修改 onShow 方法添加登录检查**

在 `publish.ts` 中修改 `onShow` 方法，添加登录状态检查和弹窗逻辑：

```typescript
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
  }
```

- [ ] **Step 2: 在微信开发者工具中测试**

手动测试场景：
1. 清除本地存储中的 token（模拟未登录状态）
2. 点击发布 tab 进入发布页
3. 验证弹窗显示正确（标题"请先登录"，内容"发布商品需要登录"）
4. 点击"取消"按钮 → 应留在发布页
5. 再次进入发布页，点击"去登录"按钮 → 应跳转到个人中心页

- [ ] **Step 3: 提交代码**

```bash
git add mall-mini/miniprogram/pages/publish/publish.ts
git commit -m "feat: add login check before publishing product"
```
