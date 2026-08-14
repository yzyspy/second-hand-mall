import request from './request'

export function searchProducts(params) {
  return request.get('/api/product/search', { params })
}
