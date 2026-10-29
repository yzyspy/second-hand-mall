package com.yzy.mall.model

import com.yzy.mall.network.ProductDto

/** 与小程序 pages/home/home.ts 及 iOS Product.swift 的映射逻辑一致 */
data class Product(
    val id: Int,
    val title: String,
    val price: Double,
    val location: String,
    val category: String,
    val seller: String,
    val avatar: String,
    val images: List<String>,
    val buyUid: Int,
    val createTime: String
) {
    val coverUrl: String? get() = images.firstOrNull()

    val priceText: String get() {
        val rounded = price.toLong()
        return if (price == rounded.toDouble()) "¥$rounded" else "¥${"%.2f".format(price)}"
    }
}

fun ProductDto.toProduct(): Product = Product(
    id = id,
    title = title,
    price = price,
    location = location.orEmpty(),
    category = category.orEmpty(),
    seller = seller?.takeIf { it.isNotEmpty() } ?: "匿名用户",
    avatar = avatar.orEmpty(),
    images = images?.split(",")?.filter { it.isNotEmpty() } ?: emptyList(),
    buyUid = buy_uid ?: 0,
    createTime = create_time.orEmpty()
)
