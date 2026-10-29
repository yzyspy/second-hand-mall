package com.yzy.mall.data

import android.content.Context
import com.google.gson.Gson
import com.google.gson.reflect.TypeToken
import java.io.BufferedReader

data class RegionItem(val name: String, val children: List<CityItem>)
data class CityItem(val name: String, val children: List<String>)

/** 复用 mall-mini 的 data/china-regions.ts 数据（assets/china_regions.json） */
object ChinaRegions {
    @Volatile
    private var cached: List<RegionItem>? = null

    fun load(context: Context): List<RegionItem> {
        cached?.let { return it }
        val json = context.assets.open("china_regions.json").bufferedReader().use(BufferedReader::readText)
        val type = object : TypeToken<List<RegionItem>>() {}.type
        val result = Gson().fromJson<List<RegionItem>>(json, type)
        cached = result
        return result
    }
}
