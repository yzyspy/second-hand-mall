export function buildRegionCascaderOptions(regionsData) {
  return regionsData.map((province) => ({
    text: province.name,
    value: province.name,
    children: province.children.map((city) => ({
      text: city.name,
      value: city.name,
      children: city.children.map((district) => ({
        text: district,
        value: district
      }))
    }))
  }))
}
