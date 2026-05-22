import request from '@/utils/request'

export interface AdminUser {
  id: number
  user_name: string
  nick_name: string
  phone: string
  avatar: string
  is_banned: boolean
  product_count: number
  created_at: string
}

export interface UserListParams {
  keyword?: string
  is_banned?: boolean
  page: number
  page_size: number
}

export function getUserList(params: UserListParams) {
  return request.get<any, { code: number; data: { list: AdminUser[]; total: number } }>(
    '/admin/users', { params }
  )
}

export function getUserDetail(id: number) {
  return request.get<any, { code: number; data: AdminUser }>(
    `/admin/users/${id}`
  )
}

export function banUser(id: number) {
  return request.post<any, { code: number; msg: string }>(`/admin/users/${id}/ban`)
}

export function unbanUser(id: number) {
  return request.post<any, { code: number; msg: string }>(`/admin/users/${id}/unban`)
}
