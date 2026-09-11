import { defineStore } from 'pinia'
import { reactive, ref } from 'vue'
import { sampleApi, type SampleQuery } from '@/api'
import type { Sample, SampleDetail } from '@/types'

// 样品列表与详情（流转轨迹 / 检测结果）
export const useSampleStore = defineStore('sample', () => {
  const list = ref<Sample[]>([])
  const total = ref(0)
  const loading = ref(false)
  const detail = ref<SampleDetail | null>(null)
  const detailLoading = ref(false)

  const query = reactive<Required<SampleQuery>>({
    keyword: '',
    status: '',
    page: 1,
    page_size: 20,
  })

  async function fetchList() {
    loading.value = true
    try {
      const params: SampleQuery = {
        page: query.page,
        page_size: query.page_size,
      }
      if (query.keyword) params.keyword = query.keyword
      if (query.status) params.status = query.status
      const page = await sampleApi.list(params)
      list.value = page.list
      total.value = page.total
    } finally {
      loading.value = false
    }
  }

  async function fetchDetail(id: number) {
    detailLoading.value = true
    try {
      detail.value = await sampleApi.detail(id)
    } finally {
      detailLoading.value = false
    }
  }

  function clearDetail() {
    detail.value = null
  }

  return {
    list,
    total,
    loading,
    detail,
    detailLoading,
    query,
    fetchList,
    fetchDetail,
    clearDetail,
  }
})
