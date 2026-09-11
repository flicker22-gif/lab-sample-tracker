<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useSampleStore } from '@/stores/sample'
import { formatDateTime } from '@/utils/datetime'
import { STATUS_LABELS, STATUS_TAG_TYPE } from '@/constants/labels'
import type { SampleStatus } from '@/types'
import SampleCreateDialog from '@/components/SampleCreateDialog.vue'

const router = useRouter()
const store = useSampleStore()
const { list, total, loading, query } = storeToRefs(store)

const createDialog = ref<InstanceType<typeof SampleCreateDialog>>()

onMounted(() => {
  store.fetchList()
})

function search() {
  query.value.page = 1
  store.fetchList()
}

function reset() {
  query.value.keyword = ''
  query.value.status = ''
  query.value.page = 1
  store.fetchList()
}

function onPageChange(p: number) {
  query.value.page = p
  store.fetchList()
}

function openDetail(id: number) {
  router.push({ name: 'sample-detail', params: { id } })
}

const statusOptions = Object.entries(STATUS_LABELS).map(([value, label]) => ({ value, label }))

function tagType(s: SampleStatus) {
  return STATUS_TAG_TYPE[s] ?? 'info'
}
</script>

<template>
  <div class="page">
    <el-card shadow="never">
      <div class="toolbar">
        <el-form inline @submit.prevent>
          <el-form-item label="关键字">
            <el-input
              v-model="query.keyword"
              placeholder="样品编号 / 名称"
              clearable
              style="width: 220px"
              @keyup.enter="search"
              @clear="search"
            />
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="query.status" placeholder="全部" clearable style="width: 140px" @change="search">
              <el-option v-for="o in statusOptions" :key="o.value" :label="o.label" :value="o.value" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :icon="'Search'" @click="search">查询</el-button>
            <el-button :icon="'RefreshLeft'" @click="reset">重置</el-button>
          </el-form-item>
        </el-form>
        <el-button type="success" :icon="'Plus'" @click="createDialog?.open()">新增样品</el-button>
      </div>

      <el-table :data="list" v-loading="loading" stripe style="width: 100%">
        <el-table-column prop="code" label="样品编号" width="180">
          <template #default="{ row }">
            <el-link type="primary" :underline="false" @click="openDetail(row.id)">{{ row.code }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="样品名称" min-width="160" show-overflow-tooltip />
        <el-table-column prop="category" label="类型" width="100" />
        <el-table-column prop="source" label="送检单位" min-width="140" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="tagType(row.status)" size="small">{{ STATUS_LABELS[row.status as SampleStatus] ?? row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="当前位置" min-width="150">
          <template #default="{ row }">
            <span>{{ row.current_location_name || '—' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="receiver_name" label="收样人" width="90" />
        <el-table-column label="收样时间" width="160">
          <template #default="{ row }">{{ formatDateTime(row.received_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="90" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row.id)">轨迹</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无样品，点击右上角「新增样品」开始收样" />
        </template>
      </el-table>

      <div class="pager">
        <el-pagination
          background
          layout="total, prev, pager, next, jumper"
          :total="total"
          :page-size="query.page_size"
          :current-page="query.page"
          @current-change="onPageChange"
        />
      </div>
    </el-card>

    <SampleCreateDialog ref="createDialog" @created="search" />
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}
.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
