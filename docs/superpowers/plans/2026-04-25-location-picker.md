# 交易地点二级选择器实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将发布页面的交易地点从文本输入框改为二级联动选择器（省/直辖市 → 区县）。

**Architecture:** 新建静态数据文件存储省市数据，修改发布页面的 WXML 模板、TS 逻辑和 WXSS 样式，实现两个级联 picker 选择器。

**Tech Stack:** TypeScript, WeChat Mini Program APIs (picker component)

---

### Task 1: 创建省市静态数据文件

**Files:**
- Create: `miniprogram/utils/region-data.ts`

- [ ] **Step 1: 创建 region-data.ts 文件**

创建包含全国省市数据的 TypeScript 文件：

```typescript
// miniprogram/utils/region-data.ts

/**
 * 省市二级数据
 * 键：省/直辖市名称
 * 值：下辖市/区县列表
 */
export const regionData: Record<string, string[]> = {
  "北京市": [
    "东城区", "西城区", "朝阳区", "丰台区", "石景山区", "海淀区",
    "门头沟区", "房山区", "通州区", "顺义区", "昌平区", "大兴区",
    "怀柔区", "平谷区", "密云区", "延庆区"
  ],
  "天津市": [
    "和平区", "河东区", "河西区", "南开区", "河北区", "红桥区",
    "东丽区", "西青区", "津南区", "北辰区", "武清区", "宝坻区",
    "滨海新区", "宁河区", "静海区", "蓟州区"
  ],
  "河北省": [
    "石家庄市", "唐山市", "秦皇岛市", "邯郸市", "邢台市", "保定市",
    "张家口市", "承德市", "沧州市", "廊坊市", "衡水市"
  ],
  "山西省": [
    "太原市", "大同市", "阳泉市", "长治市", "晋城市", "朔州市",
    "晋中市", "运城市", "忻州市", "临汾市", "吕梁市"
  ],
  "内蒙古自治区": [
    "呼和浩特市", "包头市", "乌海市", "赤峰市", "通辽市", "鄂尔多斯市",
    "呼伦贝尔市", "巴彦淖尔市", "乌兰察布市", "兴安盟", "锡林郭勒盟", "阿拉善盟"
  ],
  "辽宁省": [
    "沈阳市", "大连市", "鞍山市", "抚顺市", "本溪市", "丹东市",
    "锦州市", "营口市", "阜新市", "辽阳市", "盘锦市", "铁岭市",
    "朝阳市", "葫芦岛市"
  ],
  "吉林省": [
    "长春市", "吉林市", "四平市", "辽源市", "通化市", "白山市",
    "松原市", "白城市", "延边朝鲜族自治州"
  ],
  "黑龙江省": [
    "哈尔滨市", "齐齐哈尔市", "鸡西市", "鹤岗市", "双鸭山市", "大庆市",
    "伊春市", "佳木斯市", "七台河市", "牡丹江市", "黑河市", "绥化市",
    "大兴安岭地区"
  ],
  "上海市": [
    "黄浦区", "徐汇区", "长宁区", "静安区", "普陀区", "虹口区",
    "杨浦区", "闵行区", "宝山区", "嘉定区", "浦东新区", "金山区",
    "松江区", "青浦区", "奉贤区", "崇明区"
  ],
  "江苏省": [
    "南京市", "无锡市", "徐州市", "常州市", "苏州市", "南通市",
    "连云港市", "淮安市", "盐城市", "扬州市", "镇江市", "泰州市", "宿迁市"
  ],
  "浙江省": [
    "杭州市", "宁波市", "温州市", "嘉兴市", "湖州市", "绍兴市",
    "金华市", "衢州市", "舟山市", "台州市", "丽水市"
  ],
  "安徽省": [
    "合肥市", "芜湖市", "蚌埠市", "淮南市", "马鞍山市", "淮北市",
    "铜陵市", "安庆市", "黄山市", "滁州市", "阜阳市", "宿州市",
    "六安市", "亳州市", "池州市", "宣城市"
  ],
  "福建省": [
    "福州市", "厦门市", "莆田市", "三明市", "泉州市", "漳州市",
    "南平市", "龙岩市", "宁德市"
  ],
  "江西省": [
    "南昌市", "景德镇市", "萍乡市", "九江市", "新余市", "鹰潭市",
    "赣州市", "吉安市", "宜春市", "抚州市", "上饶市"
  ],
  "山东省": [
    "济南市", "青岛市", "淄博市", "枣庄市", "东营市", "烟台市",
    "潍坊市", "济宁市", "泰安市", "威海市", "日照市", "临沂市",
    "德州市", "聊城市", "滨州市", "菏泽市"
  ],
  "河南省": [
    "郑州市", "开封市", "洛阳市", "平顶山市", "安阳市", "鹤壁市",
    "新乡市", "焦作市", "濮阳市", "许昌市", "漯河市", "三门峡市",
    "南阳市", "商丘市", "信阳市", "周口市", "驻马店市", "济源市"
  ],
  "湖北省": [
    "武汉市", "黄石市", "十堰市", "宜昌市", "襄阳市", "鄂州市",
    "荆门市", "孝感市", "荆州市", "黄冈市", "咸宁市", "随州市",
    "恩施土家族苗族自治州", "仙桃市", "潜江市", "天门市", "神农架林区"
  ],
  "湖南省": [
    "长沙市", "株洲市", "湘潭市", "衡阳市", "邵阳市", "岳阳市",
    "常德市", "张家界市", "益阳市", "郴州市", "永州市", "怀化市",
    "娄底市", "湘西土家族苗族自治州"
  ],
  "广东省": [
    "广州市", "韶关市", "深圳市", "珠海市", "汕头市", "佛山市",
    "江门市", "湛江市", "茂名市", "肇庆市", "惠州市", "梅州市",
    "汕尾市", "河源市", "阳江市", "清远市", "东莞市", "中山市",
    "潮州市", "揭阳市", "云浮市"
  ],
  "广西壮族自治区": [
    "南宁市", "柳州市", "桂林市", "梧州市", "北海市", "防城港市",
    "钦州市", "贵港市", "玉林市", "百色市", "贺州市", "河池市",
    "来宾市", "崇左市"
  ],
  "海南省": [
    "海口市", "三亚市", "三沙市", "儋州市", "五指山市", "琼海市",
    "文昌市", "万宁市", "东方市", "定安县", "屯昌县", "澄迈县",
    "临高县", "白沙黎族自治县", "昌江黎族自治县", "乐东黎族自治县",
    "陵水黎族自治县", "保亭黎族苗族自治县", "琼中黎族苗族自治县"
  ],
  "重庆市": [
    "万州区", "涪陵区", "渝中区", "大渡口区", "江北区", "沙坪坝区",
    "九龙坡区", "南岸区", "北碚区", "綦江区", "大足区", "渝北区",
    "巴南区", "黔江区", "长寿区", "江津区", "合川区", "永川区",
    "南川区", "璧山区", "铜梁区", "潼南区", "荣昌区", "开州区",
    "梁平区", "武隆区", "城口县", "丰都县", "垫江县", "忠县",
    "云阳县", "奉节县", "巫山县", "巫溪县", "石柱土家族自治县",
    "秀山土家族苗族自治县", "酉阳土家族苗族自治县", "彭水苗族土家族自治县"
  ],
  "四川省": [
    "成都市", "自贡市", "攀枝花市", "泸州市", "德阳市", "绵阳市",
    "广元市", "遂宁市", "内江市", "乐山市", "南充市", "眉山市",
    "宜宾市", "广安市", "达州市", "雅安市", "巴中市", "资阳市",
    "阿坝藏族羌族自治州", "甘孜藏族自治州", "凉山彝族自治州"
  ],
  "贵州省": [
    "贵阳市", "六盘水市", "遵义市", "安顺市", "毕节市", "铜仁市",
    "黔西南布依族苗族自治州", "黔东南苗族侗族自治州", "黔南布依族苗族自治州"
  ],
  "云南省": [
    "昆明市", "曲靖市", "玉溪市", "保山市", "昭通市", "丽江市",
    "普洱市", "临沧市", "楚雄彝族自治州", "红河哈尼族彝族自治州",
    "文山壮族苗族自治州", "西双版纳傣族自治州", "大理白族自治州",
    "德宏傣族景颇族自治州", "怒江傈僳族自治州", "迪庆藏族自治州"
  ],
  "西藏自治区": [
    "拉萨市", "日喀则市", "昌都市", "林芝市", "山南市", "那曲市", "阿里地区"
  ],
  "陕西省": [
    "西安市", "铜川市", "宝鸡市", "咸阳市", "渭南市", "延安市",
    "汉中市", "榆林市", "安康市", "商洛市"
  ],
  "甘肃省": [
    "兰州市", "嘉峪关市", "金昌市", "白银市", "天水市", "武威市",
    "张掖市", "平凉市", "酒泉市", "庆阳市", "定西市", "陇南市",
    "临夏回族自治州", "甘南藏族自治州"
  ],
  "青海省": [
    "西宁市", "海东市", "海北藏族自治州", "黄南藏族自治州",
    "海南藏族自治州", "果洛藏族自治州", "玉树藏族自治州", "海西蒙古族藏族自治州"
  ],
  "宁夏回族自治区": [
    "银川市", "石嘴山市", "吴忠市", "固原市", "中卫市"
  ],
  "新疆维吾尔自治区": [
    "乌鲁木齐市", "克拉玛依市", "吐鲁番市", "哈密市", "昌吉回族自治州",
    "博尔塔拉蒙古自治州", "巴音郭楞蒙古自治州", "阿克苏地区", "克孜勒苏柯尔克孜自治州",
    "喀什地区", "和田地区", "伊犁哈萨克自治州", "塔城地区", "阿勒泰地区",
    "石河子市", "阿拉尔市", "图木舒克市", "五家渠市", "北屯市",
    "铁门关市", "双河市", "可克达拉市", "昆玉市", "胡杨河市"
  ],
  "香港特别行政区": [
    "中西区", "湾仔区", "东区", "南区", "油尖旺区", "深水埗区",
    "九龙城区", "黄大仙区", "观塘区", "荃湾区", "屯门区", "元朗区",
    "北区", "大埔区", "沙田区", "西贡区", "离岛区", "葵青区"
  ],
  "澳门特别行政区": [
    "花地玛堂区", "花王堂区", "望德堂区", "大堂区", "风顺堂区",
    "嘉模堂区", "路氹填海区", "圣方济各堂区"
  ],
  "台湾省": [
    "台北市", "新北市", "桃园市", "台中市", "台南市", "高雄市",
    "基隆市", "新竹市", "嘉义市", "新竹县", "苗栗县", "彰化县",
    "南投县", "云林县", "嘉义县", "屏东县", "宜兰县", "花莲县",
    "台东县", "澎湖县", "金门县", "连江县"
  ]
}

/**
 * 获取所有省份列表
 */
export function getProvinces(): string[] {
  return Object.keys(regionData)
}

/**
 * 获取指定省份的区县列表
 */
export function getDistricts(province: string): string[] {
  return regionData[province] || []
}
```

- [ ] **Step 2: 提交代码**

```bash
git add miniprogram/utils/region-data.ts
git commit -m "feat: add region data for province-district picker"
```

---

### Task 2: 修改发布页面 TypeScript 逻辑

**Files:**
- Modify: `miniprogram/pages/publish/publish.ts`

- [ ] **Step 1: 导入 region-data 模块**

在文件顶部添加导入语句：

```typescript
import { uploadToCos } from '../../utils/cos-upload'
import { regionData } from '../../utils/region-data'
```

- [ ] **Step 2: 更新 PublishData 接口**

修改接口定义，添加省市选择相关字段，移除不再需要的字段：

```typescript
interface PublishData {
  images: UploadedImage[]
  maxImages: number
  description: string
  price: string
  location: string
  categoryIndex: number
  categories: string[]
  submitting: boolean
  // 新增：省市选择器字段
  provinces: string[]
  provinceIndex: number
  districts: string[]
  districtIndex: number
  selectedProvince: string
}
```

- [ ] **Step 3: 更新 data 初始化**

修改 `Page()` 中的 data 对象：

```typescript
Page<PublishData, WechatMiniprogram.IAnyObject>({
  data: {
    images: [],
    maxImages: 9,
    description: '',
    price: '',
    location: '',
    categoryIndex: 0,
    categories: ['电子产品', '服装鞋帽', '图书文具', '生活用品', '数码配件', '其他'],
    submitting: false,
    // 新增字段
    provinces: [],
    provinceIndex: -1,
    districts: [],
    districtIndex: -1,
    selectedProvince: ''
  },
```

- [ ] **Step 4: 更新 onLoad 方法**

修改 `onLoad` 方法，初始化省份列表：

```typescript
  onLoad() {
    const provinces = Object.keys(regionData)
    this.setData({ provinces })
  },
```

- [ ] **Step 5: 添加省份选择处理方法**

添加 `onProvinceChange` 方法：

```typescript
  // 选择省份
  onProvinceChange(e: WechatMiniprogram.PickerChange) {
    const index = Number(e.detail.value)
    const province = this.data.provinces[index]
    const districts = regionData[province] || []
    this.setData({
      provinceIndex: index,
      selectedProvince: province,
      districts,
      districtIndex: -1,
      location: '' // 重置地点
    })
  },
```

- [ ] **Step 6: 添加区县选择处理方法**

添加 `onDistrictChange` 方法：

```typescript
  // 选择区县
  onDistrictChange(e: WechatMiniprogram.PickerChange) {
    const index = Number(e.detail.value)
    const district = this.data.districts[index]
    const location = `${this.data.selectedProvince}${district}`
    this.setData({ districtIndex: index, location })
  },
```

- [ ] **Step 7: 删除旧的位置相关方法**

删除以下方法：
- `onLocationInput` 方法（约第108-110行）
- `getLocation` 方法（约第118-127行）

- [ ] **Step 8: 更新提交后清空表单逻辑**

修改 `submitForm` 方法中的清空表单部分，添加新字段的初始化：

```typescript
      // 清空表单
      setTimeout(() => {
        this.setData({
          images: [],
          description: '',
          price: '',
          location: '',
          categoryIndex: 0,
          provinceIndex: -1,
          districts: [],
          districtIndex: -1,
          selectedProvince: ''
        })
      }, 1500)
```

- [ ] **Step 9: 提交代码**

```bash
git add miniprogram/pages/publish/publish.ts
git commit -m "feat: update publish page logic for location picker"
```

---

### Task 3: 修改发布页面 WXML 模板

**Files:**
- Modify: `miniprogram/pages/publish/publish.wxml`

- [ ] **Step 1: 替换交易地点区域**

找到约第79-93行的交易地点区域，替换为：

```xml
  <!-- 地点 -->
  <view class="form-section card">
    <view class="section-title">交易地点</view>
    <view class="location-picker-group">
      <!-- 省份选择器 -->
      <picker
        mode="selector"
        range="{{provinces}}"
        value="{{provinceIndex >= 0 ? provinceIndex : 0}}"
        bindchange="onProvinceChange"
      >
        <view class="picker-display">
          <text>{{provinceIndex >= 0 ? provinces[provinceIndex] : '请选择省份'}}</text>
          <text class="picker-arrow">›</text>
        </view>
      </picker>
      <!-- 区县选择器 -->
      <picker
        mode="selector"
        range="{{districts}}"
        value="{{districtIndex >= 0 ? districtIndex : 0}}"
        bindchange="onDistrictChange"
        disabled="{{!selectedProvince}}"
      >
        <view class="picker-display {{!selectedProvince ? 'disabled' : ''}}">
          <text>{{districtIndex >= 0 ? districts[districtIndex] : '请选择区县'}}</text>
          <text class="picker-arrow">›</text>
        </view>
      </picker>
    </view>
  </view>
```

- [ ] **Step 2: 提交代码**

```bash
git add miniprogram/pages/publish/publish.wxml
git commit -m "feat: replace location input with province-district pickers"
```

---

### Task 4: 修改发布页面 WXSS 样式

**Files:**
- Modify: `miniprogram/pages/publish/publish.wxss`

- [ ] **Step 1: 删除旧的地点输入框样式**

找到约第155-170行，删除 `.location-input-wrapper` 和 `.location-input` 和 `.location-btn` 样式：

```css
/* 删除以下样式 */
.location-input-wrapper {
  display: flex;
  align-items: center;
}

.location-input {
  flex: 1;
  font-size: 28rpx;
}

.location-btn {
  padding: 10rpx 16rpx;
  background: #f0f4f7;
  border-radius: 8rpx;
}
```

- [ ] **Step 2: 添加地点选择器组样式**

在删除的位置添加新样式：

```css
/* 地点选择器组 */
.location-picker-group {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.location-picker-group .picker-display {
  padding: 18rpx 0;
}

.location-picker-group .picker-display.disabled {
  opacity: 0.5;
}

.location-picker-group .picker-display.disabled text {
  color: #c6ced6;
}
```

- [ ] **Step 3: 提交代码**

```bash
git add miniprogram/pages/publish/publish.wxss
git commit -m "feat: update styles for location picker group"
```

---

### Task 5: 手动验证

**Files:**
- Test: 在微信开发者工具中测试

- [ ] **Step 1: 在微信开发者工具中打开项目**

打开 `mall-mini` 项目，编译运行。

- [ ] **Step 2: 测试省份选择**

1. 进入发布页面
2. 点击"交易地点"区域的第一个选择器
3. 验证省份列表正常显示
4. 选择任意省份
5. 验证区县选择器从灰显变为可用

- [ ] **Step 3: 测试区县选择**

1. 在已选省份的情况下，点击区县选择器
2. 验证区县列表正常显示
3. 选择任意区县
4. 验证地点已正确拼接（如"广东省深圳市"）

- [ ] **Step 4: 测试联动重置**

1. 重新选择不同省份
2. 验证之前选择的区县被重置
3. 验证 location 字段被清空

- [ ] **Step 5: 测试表单提交**

1. 填写完整表单（图片、描述、价格、分类、地点）
2. 点击发布
3. 验证提交成功后表单清空，地点选择器恢复初始状态

- [ ] **Step 6: 最终提交**

```bash
git status
```

确认所有修改已提交。

---

## 完成检查

- [ ] 所有文件修改完成
- [ ] 手动测试通过
- [ ] 代码已提交
