export function buildSearchParams({ page, category, province, city, district }) {
  const params = {
    page,
    page_size: 10,
    status: 0
  }
  if (category) params.category = category
  if (province) params.province = province
  if (city) params.city = city
  if (district) params.district = district
  return params
}
