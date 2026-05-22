import request from '@/utils/request'

export interface AdminProduct {
  id: number
  title: string
  price: number
  images: string
  location: string
  status: number
  category: string
  seller: string
  create_time: string
}

export interface ProductListParams {
  keyword?: string
  category?: string
  province?: string
  status?: number
  page: number
  page_size: number
}

export function getProductList(params: ProductListParams) {
  return request.get<any, { code: number; data: { list: AdminProduct[]; total: number } }>(
    '/admin/products', { params }
  )
}

export function getProductDetail(id: number) {
  return request.get<any, { code: number; data: AdminProduct }>(`/admin/products/${id}`)
}

export function delistProduct(id: number) {
  return request.post<any, { code: number; msg: string }>(`/admin/products/${id}/delist`)
}
