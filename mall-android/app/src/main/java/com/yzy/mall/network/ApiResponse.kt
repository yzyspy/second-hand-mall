package com.yzy.mall.network

data class ApiResponse<T>(
    val code: Int,
    val msg: String,
    val data: T?
)

data class ProductPage(
    val list: List<ProductDto>,
    val total: Int,
    val page: Int,
    val page_size: Int
)
