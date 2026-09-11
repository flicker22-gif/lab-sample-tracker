import { defineStore } from 'pinia'
import { ref } from 'vue'
import { metaApi } from '@/api'
import type { Location, User } from '@/types'

// 操作人员与存放位置的基础数据（下拉选项）
export const useMetaStore = defineStore('meta', () => {
  const users = ref<User[]>([])
  const locations = ref<Location[]>([])
  const loaded = ref(false)
  const loading = ref(false)

  async function load(force = false) {
    if ((loaded.value && !force) || loading.value) return
    loading.value = true
    try {
      const [u, l] = await Promise.all([metaApi.users(), metaApi.locations()])
      users.value = u
      locations.value = l
      loaded.value = true
    } finally {
      loading.value = false
    }
  }

  function userName(id: number): string {
    return users.value.find((u) => u.id === id)?.full_name ?? String(id)
  }

  function locationName(id: number | null | undefined): string {
    if (id == null) return '—'
    return locations.value.find((l) => l.id === id)?.name ?? `#${id}`
  }

  return { users, locations, loaded, load, userName, locationName }
})
