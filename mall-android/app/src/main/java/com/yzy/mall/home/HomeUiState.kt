package com.yzy.mall.home

import com.yzy.mall.data.RegionItem
import com.yzy.mall.model.Product

val CATEGORIES = listOf("电子产品", "服装鞋帽", "图书文具", "生活用品", "数码配件", "其他")

data class HomeUiState(
    val products: List<Product> = emptyList(),
    val isRefreshing: Boolean = false,
    val isLoadingMore: Boolean = false,
    val initialLoaded: Boolean = false,
    val hasMore: Boolean = true,
    val errorMessage: String? = null,

    val keyword: String = "",
    val selectedCategory: String = "",
    val selectedProvince: String = "",
    val selectedCity: String = "",
    val selectedDistrict: String = "",

    val showCategoryPanel: Boolean = false,
    val showRegionPanel: Boolean = false,
    val regionStep: Int = 0,
    val regions: List<RegionItem> = emptyList(),
    val regionCities: List<String> = emptyList(),
    val regionDistricts: List<String> = emptyList(),
) {
    val regionLabel: String
        get() = selectedDistrict.ifEmpty { selectedCity.ifEmpty { selectedProvince } }
}
