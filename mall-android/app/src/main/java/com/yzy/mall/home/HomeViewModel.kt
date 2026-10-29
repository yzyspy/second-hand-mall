package com.yzy.mall.home

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.yzy.mall.data.ChinaRegions
import com.yzy.mall.model.toProduct
import com.yzy.mall.network.RetrofitClient
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

private const val PAGE_SIZE = 10

class HomeViewModel(application: Application) : AndroidViewModel(application) {

    private val _uiState = MutableStateFlow(HomeUiState())
    val uiState: StateFlow<HomeUiState> = _uiState

    private var page = 1
    private var loadJob: kotlinx.coroutines.Job? = null

    init {
        _uiState.update { it.copy(regions = ChinaRegions.load(application)) }
        refresh()
    }

    fun refresh() {
        page = 1
        load(isRefresh = true)
    }

    fun loadMore() {
        val state = _uiState.value
        if (!state.hasMore || state.isLoadingMore || state.isRefreshing) return
        load(isRefresh = false)
    }

    private fun load(isRefresh: Boolean) {
        loadJob?.cancel()
        loadJob = viewModelScope.launch {
            _uiState.update {
                if (isRefresh) it.copy(isRefreshing = true, errorMessage = null)
                else it.copy(isLoadingMore = true)
            }
            val state = _uiState.value
            try {
                val response = RetrofitClient.apiService.searchProducts(
                    page = page,
                    pageSize = PAGE_SIZE,
                    keyword = state.keyword.ifEmpty { null },
                    category = state.selectedCategory.ifEmpty { null },
                    province = state.selectedProvince.ifEmpty { null },
                    city = state.selectedCity.ifEmpty { null },
                    district = state.selectedDistrict.ifEmpty { null },
                )
                if (response.code == 0 && response.data != null) {
                    val fetched = response.data.list.map { it.toProduct() }
                    page += 1
                    _uiState.update {
                        it.copy(
                            products = if (isRefresh) fetched else it.products + fetched,
                            isRefreshing = false,
                            isLoadingMore = false,
                            initialLoaded = true,
                            hasMore = fetched.size >= PAGE_SIZE,
                        )
                    }
                } else {
                    _uiState.update {
                        it.copy(isRefreshing = false, isLoadingMore = false, initialLoaded = true, errorMessage = response.msg)
                    }
                }
            } catch (e: Exception) {
                _uiState.update {
                    it.copy(
                        isRefreshing = false,
                        isLoadingMore = false,
                        initialLoaded = true,
                        errorMessage = e.message ?: "加载失败",
                        products = if (isRefresh) emptyList() else it.products,
                    )
                }
            }
        }
    }

    fun onKeywordChange(value: String) {
        _uiState.update { it.copy(keyword = value) }
    }

    fun onSearchSubmit() {
        refresh()
    }

    fun openCategoryPanel() = _uiState.update { it.copy(showCategoryPanel = true) }
    fun closeCategoryPanel() = _uiState.update { it.copy(showCategoryPanel = false) }

    fun selectCategory(value: String) {
        _uiState.update { it.copy(selectedCategory = value, showCategoryPanel = false) }
        refresh()
    }

    fun clearCategory() {
        _uiState.update { it.copy(selectedCategory = "") }
        refresh()
    }

    fun openRegionPanel() = _uiState.update { it.copy(showRegionPanel = true, regionStep = 0) }
    fun closeRegionPanel() = _uiState.update { it.copy(showRegionPanel = false) }

    fun selectProvince(index: Int) {
        val province = _uiState.value.regions.getOrNull(index) ?: return
        _uiState.update {
            it.copy(
                selectedProvince = province.name,
                selectedCity = "",
                selectedDistrict = "",
                regionCities = province.children.map { c -> c.name },
                regionStep = 1,
            )
        }
    }

    fun selectCity(index: Int) {
        val state = _uiState.value
        val province = state.regions.find { it.name == state.selectedProvince } ?: return
        val city = province.children.getOrNull(index) ?: return
        _uiState.update {
            it.copy(
                selectedCity = city.name,
                selectedDistrict = "",
                regionDistricts = city.children,
                regionStep = 2,
            )
        }
    }

    fun selectDistrict(district: String) {
        _uiState.update { it.copy(selectedDistrict = district, showRegionPanel = false) }
        refresh()
    }

    fun confirmRegion() {
        _uiState.update { it.copy(showRegionPanel = false) }
        refresh()
    }

    fun clearRegion() {
        _uiState.update {
            it.copy(selectedProvince = "", selectedCity = "", selectedDistrict = "", regionCities = emptyList(), regionDistricts = emptyList())
        }
        refresh()
    }
}
