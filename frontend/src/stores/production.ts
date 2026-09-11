import { defineStore } from 'pinia'
import { computed, reactive, ref } from 'vue'
import { productionApi, type LotQuery } from '@/api/production'
import type { LotDetail, Machine, Process, WaferLot } from '@/types/production'

// 工艺、机台、晶圆批次
export const useProductionStore = defineStore('production', () => {
  const processes = ref<Process[]>([])
  const machines = ref<Machine[]>([])
  const metaLoaded = ref(false)

  const list = ref<WaferLot[]>([])
  const total = ref(0)
  const loading = ref(false)
  const detail = ref<LotDetail | null>(null)
  const detailLoading = ref(false)

  const query = reactive<Required<LotQuery>>({
    keyword: '',
    status: '',
    page: 1,
    page_size: 20,
  })

  // 机台按工艺分组，便于看板展示
  const machinesByProcess = computed(() => {
    const map = new Map<number, Machine[]>()
    for (const m of machines.value) {
      const arr = map.get(m.process_id) ?? []
      arr.push(m)
      map.set(m.process_id, arr)
    }
    return map
  })

  async function loadMeta(force = false) {
    if (metaLoaded.value && !force) return
    const [ps, ms] = await Promise.all([productionApi.processes(), productionApi.machines()])
    processes.value = ps
    machines.value = ms
    metaLoaded.value = true
  }

  async function fetchMachines() {
    machines.value = await productionApi.machines()
  }

  async function fetchLots() {
    loading.value = true
    try {
      const params: LotQuery = { page: query.page, page_size: query.page_size }
      if (query.keyword) params.keyword = query.keyword
      if (query.status) params.status = query.status
      const page = await productionApi.listLots(params)
      list.value = page.list
      total.value = page.total
    } finally {
      loading.value = false
    }
  }

  async function fetchDetail(id: number) {
    detailLoading.value = true
    try {
      detail.value = await productionApi.getLot(id)
    } finally {
      detailLoading.value = false
    }
  }

  function clearDetail() {
    detail.value = null
  }

  return {
    processes,
    machines,
    metaLoaded,
    machinesByProcess,
    list,
    total,
    loading,
    detail,
    detailLoading,
    query,
    loadMeta,
    fetchMachines,
    fetchLots,
    fetchDetail,
    clearDetail,
  }
})
