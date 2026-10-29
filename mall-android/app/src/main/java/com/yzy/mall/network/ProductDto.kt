package com.yzy.mall.network

/** 对应 mall-server GET /api/product/search 返回的原始商品项（dao.ProductSearchResult）*/
data class ProductDto(
    val id: Int,
    val title: String,
    val price: Double,
    val location: String?,
    val category: String?,
    val seller: String?,
    val avatar: String?,
    val images: String?,
    val buy_uid: Int?,
    val create_time: String?
)
