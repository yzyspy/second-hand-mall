import { request } from '../../utils/request'

interface MessageItem {
  id: number
  sender_id: number
  content: string
  created_at: string
}

interface SendResp {
  conversation_id: number
  message_id: number
}

Page({
  data: {
    messages: [] as MessageItem[],
    inputContent: '',
    conversationId: 0,
    productId: 0,
    receiverId: 0,
    myId: 0,
    lastId: 0
  },

  pollTimer: null as ReturnType<typeof setInterval> | null,

  onLoad(query: Record<string, string>) {
    const myId = wx.getStorageSync('userId') as number
    const productId = Number(query['product_id'] || 0)
    const receiverId = Number(query['receiver_id'] || 0)
    this.setData({ myId, productId, receiverId })

    if (query['conversation_id']) {
      const convId = Number(query['conversation_id'])
      this.setData({ conversationId: convId })
      this.loadMessages(convId)
      request({ url: `/api/chat/read/${convId}`, method: 'PUT' }).catch(() => {})
      this.startPolling(convId)
    }
  },

  onUnload() {
    if (this.pollTimer) clearInterval(this.pollTimer)
  },

  startPolling(convId: number) {
    this.pollTimer = setInterval(() => {
      this.loadMessages(convId)
    }, 3000)
  },

  loadMessages(convId: number) {
    const lastId = this.data.lastId
    request<MessageItem[]>({ url: `/api/chat/messages?conv_id=${convId}&last_id=${lastId}` }).then(resp => {
      const data = resp.data
      if (!data || data.length === 0) return
      const newLastId = data[data.length - 1].id
      this.setData({
        messages: [...this.data.messages, ...data],
        lastId: newLastId
      })
    })
  },

  onInputChange(e: WechatMiniprogram.Input) {
    this.setData({ inputContent: e.detail.value })
  },

  onSend() {
    const content = this.data.inputContent.trim()
    if (!content) return

    request<SendResp>({
      url: '/api/chat/send',
      method: 'POST',
      data: {
        product_id: this.data.productId,
        receiver_id: this.data.receiverId,
        content
      }
    }).then(resp => {
      const sendData = resp.data
      if (!sendData) return
      const convId = sendData.conversation_id
      this.setData({ conversationId: convId, inputContent: '' })
      if (!this.pollTimer) {
        this.loadMessages(convId)
        this.startPolling(convId)
        request({ url: `/api/chat/read/${convId}`, method: 'PUT' }).catch(() => {})
      }
    }).catch(() => {
      wx.showToast({ title: '发送失败', icon: 'none' })
    })
  }
})
