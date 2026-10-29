package com.yzy.mall.network

import retrofit2.http.GET
import retrofit2.http.Query

interface ApiService {

    /** GET /api/product/search（公开接口） */
    @GET("api/product/search")
    suspend fun searchProducts(
        @Query("page") page: Int,
        @Query("page_size") pageSize: Int,
        @Query("status") status: Int = 0,
        @Query("keyword") keyword: String? = null,
        @Query("category") category: String? = null,
        @Query("province") province: String? = null,
        @Query("city") city: String? = null,
        @Query("district") district: String? = null
    ): ApiResponse<ProductPage>
}
