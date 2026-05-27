import { request } from '../../utils/request'

interface ConversationItem {
  conversation_id: number
  product: { id: number; title: string; cover: string }
  other_user: { id: number; nickname: string; avatar: string }
  last_message: string
  last_at: string
  unread_count: number
}

Page({
  data: {
    conversations: [] as ConversationItem[]
  },

  onShow() {
    this.loadConversations()
  },

  loadConversations() {
    request<ConversationItem[]>({ url: '/api/chat/conversations' }).then(resp => {
      this.setData({ conversations: resp.data || [] })
    }).catch(() => {
      wx.showToast({ title: '加载失败', icon: 'none' })
    })
  },

  goToChat(e: WechatMiniprogram.TouchEvent) {
    const { convId, productId, receiverId } = e.currentTarget.dataset as {
      convId: number; productId: number; receiverId: number
    }
    wx.navigateTo({
      url: `/pages/chat/chat?conversation_id=${convId}&product_id=${productId}&receiver_id=${receiverId}`
    })
  }
})
