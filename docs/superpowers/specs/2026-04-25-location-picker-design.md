# 交易地点二级选择器设计

## 概述

将发布页面的交易地点从文本输入框改为二级联动选择器：先选择省/直辖市，再选择区县。

## 数据层

### 静态数据文件

**路径:** `miniprogram/utils/region-data.ts`

**数据结构:**
```typescript
export const regionData: Record<string, string[]> = {
  "北京市": ["东城区", "西城区", "朝阳区", "丰台区", "石景山区", "海淀区", ...],
  "上海市": ["黄浦区", "徐汇区", "长宁区", "静安区", "普陀区", ...],
  "广东省": ["广州市", "深圳市", "珠海市", "汕头市", "佛山市", ...],
  // ... 其他省份
}
```

**数据范围:** 覆盖全国省、自治区、直辖市及特别行政区，每个省下辖主要城市/区县。

## UI层

### 组件结构

**文件:** `miniprogram/pages/publish/publish.wxml`

**修改点:**
- 移除现有文本输入框 `<input class="location-input">` 和 GPS 定位按钮
- 替换为两个 picker 选择器：

```xml
<view class="form-section card">
  <view class="section-title">交易地点</view>
  <view class="location-picker-group">
    <!-- 省份选择器 -->
    <picker
      mode="selector"
      range="{{provinces}}"
      value="{{provinceIndex}}"
      bindchange="onProvinceChange"
    >
      <view class="picker-display">
        <text>{{provinces[provinceIndex] || '请选择省份'}}</text>
        <text class="picker-arrow">›</text>
      </view>
    </picker>
    <!-- 区县选择器 -->
    <picker
      mode="selector"
      range="{{districts}}"
      value="{{districtIndex}}"
      bindchange="onDistrictChange"
      disabled="{{!selectedProvince}}"
    >
      <view class="picker-display {{!selectedProvince ? 'disabled' : ''}}">
        <text>{{districts[districtIndex] || '请选择区县'}}</text>
        <text class="picker-arrow">›</text>
      </view>
    </picker>
  </view>
</view>
```

### 样式

**文件:** `miniprogram/pages/publish/publish.wxss`

新增 `.location-picker-group` 布局样式，两个 picker 水平排列或垂直堆叠（移动端建议垂直堆叠）。

## 逻辑层

### 页面数据

**文件:** `miniprogram/pages/publish/publish.ts`

```typescript
interface PublishData {
  // ... 现有字段
  provinces: string[]           // 省份列表
  provinceIndex: number         // 当前选中的省份索引
  districts: string[]           // 当前省份下的区县列表
  districtIndex: number         // 当前选中的区县索引
  selectedProvince: string      // 已选省份名称
  location: string              // 最终拼接的地点文本
}
```

### 方法

```typescript
// 初始化数据
onLoad() {
  const provinces = Object.keys(regionData)
  this.setData({ provinces })
}

// 省份变更
onProvinceChange(e: WechatMiniprogram.PickerChange) {
  const index = Number(e.detail.value)
  const province = this.data.provinces[index]
  const districts = regionData[province] || []
  this.setData({
    provinceIndex: index,
    selectedProvince: province,
    districts,
    districtIndex: 0,
    location: '' // 重置地点
  })
}

// 区县变更
onDistrictChange(e: WechatMiniprogram.PickerChange) {
  const index = Number(e.detail.value)
  const district = this.data.districts[index]
  const location = `${this.data.selectedProvince}${district}`
  this.setData({ districtIndex: index, location })
}
```

## 存储层

### 数据格式

`location` 字段存储拼接后的完整文本，例如：
- `"北京市朝阳区"`
- `"广东省深圳市"`

### 数据库

无需修改 schema，`location` 字段仍为 `varchar(100)` 字符串类型。

## 验证逻辑

更新 `validateForm()` 方法：
- 检查 `location` 非空（必须完成二级选择）

## 用户体验

1. 进入发布页，省份选择器显示"请选择省份"占位文本
2. 用户点击省份 picker，选择后区县 picker 激活
3. 用户选择区县后，`location` 自动拼接显示
4. 未选省份时，区县 picker 灰显不可点

## 文件变更清单

| 文件 | 操作 |
|------|------|
| `miniprogram/utils/region-data.ts` | 新建 |
| `miniprogram/pages/publish/publish.ts` | 修改 |
| `miniprogram/pages/publish/publish.wxml` | 修改 |
| `miniprogram/pages/publish/publish.wxss` | 修改 |
