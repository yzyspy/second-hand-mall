import request from '@/utils/request'

export function adminLogin(username: string, password: string) {
  return request.post<any, { code: number; msg: string; data: { token: string; username: string } }>(
    '/admin/login',
    { username, password }
  )
}
