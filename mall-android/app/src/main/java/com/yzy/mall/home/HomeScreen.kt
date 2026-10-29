package com.yzy.mall.home

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.LocationOn
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.AssistChip
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import coil.compose.AsyncImage
import com.yzy.mall.model.Product

@Composable
fun HomeScreen(modifier: Modifier = Modifier, viewModel: HomeViewModel = viewModel()) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    Column(modifier = modifier.fillMaxSize()) {
        SearchBar(
            keyword = uiState.keyword,
            onKeywordChange = viewModel::onKeywordChange,
            onSearch = viewModel::onSearchSubmit,
        )
        FilterChipRow(
            selectedCategory = uiState.selectedCategory,
            regionLabel = uiState.regionLabel,
            onOpenCategoryPanel = viewModel::openCategoryPanel,
            onClearCategory = viewModel::clearCategory,
            onOpenRegionPanel = viewModel::openRegionPanel,
            onClearRegion = viewModel::clearRegion,
        )
        ProductList(
            uiState = uiState,
            onRefresh = viewModel::refresh,
            onLoadMore = viewModel::loadMore,
        )
    }

    if (uiState.showCategoryPanel) {
        CategoryPanel(
            selectedCategory = uiState.selectedCategory,
            onDismiss = viewModel::closeCategoryPanel,
            onSelect = viewModel::selectCategory,
        )
    }

    if (uiState.showRegionPanel) {
        RegionPanel(
            uiState = uiState,
            onDismiss = viewModel::closeRegionPanel,
            onSelectProvince = viewModel::selectProvince,
            onSelectCity = viewModel::selectCity,
            onSelectDistrict = viewModel::selectDistrict,
            onConfirm = viewModel::confirmRegion,
        )
    }
}

@Composable
private fun SearchBar(keyword: String, onKeywordChange: (String) -> Unit, onSearch: () -> Unit) {
    OutlinedTextField(
        value = keyword,
        onValueChange = onKeywordChange,
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 12.dp, vertical = 8.dp),
        placeholder = { Text("搜索商品") },
        leadingIcon = { Icon(Icons.Filled.Search, contentDescription = null) },
        singleLine = true,
        shape = RoundedCornerShape(24.dp),
        keyboardActions = androidx.compose.foundation.text.KeyboardActions(onSearch = { onSearch() }),
        keyboardOptions = androidx.compose.foundation.text.KeyboardOptions(
            imeAction = androidx.compose.ui.text.input.ImeAction.Search
        ),
    )
}

@Composable
private fun FilterChipRow(
    selectedCategory: String,
    regionLabel: String,
    onOpenCategoryPanel: () -> Unit,
    onClearCategory: () -> Unit,
    onOpenRegionPanel: () -> Unit,
    onClearRegion: () -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 12.dp, vertical = 4.dp),
        horizontalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        AssistChip(
            onClick = onOpenCategoryPanel,
            label = { Text(selectedCategory.ifEmpty { "全部分类" }) },
            trailingIcon = {
                if (selectedCategory.isNotEmpty()) {
                    Text("✕", modifier = Modifier.clickable { onClearCategory() })
                } else {
                    Text("▾")
                }
            },
        )
        AssistChip(
            onClick = onOpenRegionPanel,
            label = { Text(regionLabel.ifEmpty { "全部地区" }) },
            trailingIcon = {
                if (regionLabel.isNotEmpty()) {
                    Text("✕", modifier = Modifier.clickable { onClearRegion() })
                } else {
                    Text("▾")
                }
            },
        )
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun ProductList(
    uiState: HomeUiState,
    onRefresh: () -> Unit,
    onLoadMore: () -> Unit,
) {
    val listState = rememberLazyListState()
    val shouldLoadMore by remember {
        derivedStateOf {
            val lastVisible = listState.layoutInfo.visibleItemsInfo.lastOrNull()?.index ?: 0
            lastVisible >= listState.layoutInfo.totalItemsCount - 3
        }
    }
    androidx.compose.runtime.LaunchedEffect(shouldLoadMore, uiState.hasMore, uiState.products.size) {
        if (shouldLoadMore && uiState.products.isNotEmpty()) onLoadMore()
    }

    PullToRefreshBox(
        isRefreshing = uiState.isRefreshing,
        onRefresh = onRefresh,
        modifier = Modifier.fillMaxSize(),
    ) {
        when {
            uiState.errorMessage != null && uiState.products.isEmpty() -> {
                EmptyState(text = uiState.errorMessage, showRetry = true, onRetry = onRefresh)
            }
            uiState.products.isEmpty() && uiState.initialLoaded && !uiState.isRefreshing -> {
                EmptyState(text = "暂无相关商品", showRetry = false, onRetry = onRefresh)
            }
            else -> {
                LazyColumn(
                    state = listState,
                    contentPadding = androidx.compose.foundation.layout.PaddingValues(12.dp),
                    verticalArrangement = Arrangement.spacedBy(12.dp),
                    modifier = Modifier.fillMaxSize(),
                ) {
                    itemsIndexed(uiState.products, key = { _, item -> item.id }) { _, product ->
                        ProductCard(product)
                    }
                    if (uiState.isLoadingMore) {
                        item {
                            Box(modifier = Modifier.fillMaxWidth(), contentAlignment = Alignment.Center) {
                                CircularProgressIndicator(modifier = Modifier.padding(16.dp))
                            }
                        }
                    }
                    if (!uiState.hasMore && uiState.products.isNotEmpty()) {
                        item {
                            Text(
                                "已经到底了~",
                                modifier = Modifier.fillMaxWidth().padding(16.dp),
                                textAlign = androidx.compose.ui.text.style.TextAlign.Center,
                                color = MaterialTheme.colorScheme.onSurfaceVariant,
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun EmptyState(text: String, showRetry: Boolean, onRetry: () -> Unit) {
    Column(
        modifier = Modifier.fillMaxSize().padding(top = 100.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text(text, color = MaterialTheme.colorScheme.onSurfaceVariant)
        if (showRetry) {
            Spacer(modifier = Modifier.height(12.dp))
            Button(onClick = onRetry) { Text("重试") }
        }
    }
}

@Composable
private fun ProductCard(product: Product) {
    Surface(
        shape = RoundedCornerShape(12.dp),
        color = MaterialTheme.colorScheme.surface,
        modifier = Modifier.fillMaxWidth(),
    ) {
        Row(modifier = Modifier.padding(12.dp)) {
            Box(modifier = Modifier.size(110.dp)) {
                AsyncImage(
                    model = product.coverUrl,
                    contentDescription = product.title,
                    contentScale = ContentScale.Crop,
                    modifier = Modifier
                        .fillMaxSize()
                        .clip(RoundedCornerShape(8.dp)),
                )
                Surface(
                    color = MaterialTheme.colorScheme.primary,
                    shape = RoundedCornerShape(50),
                    modifier = Modifier
                        .align(Alignment.BottomStart)
                        .padding(6.dp),
                ) {
                    Text(
                        product.priceText,
                        color = Color.White,
                        fontWeight = FontWeight.Bold,
                        modifier = Modifier.padding(horizontal = 10.dp, vertical = 4.dp),
                    )
                }
            }
            Spacer(modifier = Modifier.width(12.dp))
            Column(modifier = Modifier.padding(top = 4.dp)) {
                Text(
                    product.title,
                    fontWeight = FontWeight.Bold,
                    style = MaterialTheme.typography.titleMedium,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                Spacer(modifier = Modifier.height(8.dp))
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(
                        Icons.Filled.LocationOn,
                        contentDescription = null,
                        tint = MaterialTheme.colorScheme.error,
                        modifier = Modifier.size(14.dp),
                    )
                    Text(
                        " ${product.location}  卖家：${product.seller}",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                    )
                }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun CategoryPanel(
    selectedCategory: String,
    onDismiss: () -> Unit,
    onSelect: (String) -> Unit,
) {
    ModalBottomSheet(onDismissRequest = onDismiss) {
        Column(modifier = Modifier.padding(bottom = 24.dp)) {
            PanelItem(text = "全部", selected = selectedCategory.isEmpty()) { onSelect("") }
            CATEGORIES.forEach { category ->
                PanelItem(text = category, selected = category == selectedCategory) { onSelect(category) }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun RegionPanel(
    uiState: HomeUiState,
    onDismiss: () -> Unit,
    onSelectProvince: (Int) -> Unit,
    onSelectCity: (Int) -> Unit,
    onSelectDistrict: (String) -> Unit,
    onConfirm: () -> Unit,
) {
    ModalBottomSheet(onDismissRequest = onDismiss, sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true)) {
        Column {
            Row(
                modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 8.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                val breadcrumb = uiState.selectedProvince.ifEmpty { "请选择省份" } +
                    if (uiState.selectedCity.isNotEmpty()) " / ${uiState.selectedCity}" else ""
                Text(breadcrumb, style = MaterialTheme.typography.titleSmall)
                if (uiState.regionStep > 0) {
                    Text(
                        "确认",
                        color = MaterialTheme.colorScheme.primary,
                        fontWeight = FontWeight.Bold,
                        modifier = Modifier.clickable { onConfirm() },
                    )
                }
            }
            LazyColumn(modifier = Modifier.height(360.dp)) {
                when (uiState.regionStep) {
                    0 -> itemsIndexed(uiState.regions) { index, region ->
                        PanelItem(text = region.name, selected = region.name == uiState.selectedProvince) {
                            onSelectProvince(index)
                        }
                    }
                    1 -> itemsIndexed(uiState.regionCities) { index, city ->
                        PanelItem(text = city, selected = city == uiState.selectedCity) {
                            onSelectCity(index)
                        }
                    }
                    else -> itemsIndexed(uiState.regionDistricts) { _, district ->
                        PanelItem(text = district, selected = district == uiState.selectedDistrict) {
                            onSelectDistrict(district)
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun PanelItem(text: String, selected: Boolean, onClick: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .background(if (selected) MaterialTheme.colorScheme.primaryContainer else Color.Transparent)
            .padding(horizontal = 16.dp, vertical = 14.dp),
    ) {
        Text(
            text,
            color = if (selected) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.onSurface,
            fontWeight = if (selected) FontWeight.Bold else FontWeight.Normal,
        )
    }
}
