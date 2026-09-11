<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useProductionStore } from '@/stores/production'
import { formatDateTime } from '@/utils/datetime'
import { LOT_STATUS_LABELS, LOT_STATUS_TAG } from '@/constants/production'
import type { LotStatus } from '@/types/production'
import LotCreateDialog from '@/components/LotCreateDialog.vue'

const router = useRouter()
const store = useProductionStore()
const { list, total, loading, query } = storeToRefs(store)

const createDialog = ref<InstanceType<typeof LotCreateDialog>>()

onMounted(() => {
  store.fetchLots()
})

function search() {
  query.value.page = 1
  store.fetchLots()
}
function reset() {
  query.value.keyword = ''
  query.value.status = ''
  query.value.page = 1
  store.fetchLots()
}
function onPageChange(p: number) {
  query.value.page = p
  store.fetchLots()
}
function openDetail(id: number) {
  router.push({ name: 'lot-detail', params: { id } })
}

const statusOptions = Object.entries(LOT_STATUS_LABELS).map(([value, label]) => ({ value, label }))
</script>

<template>
  <div class="page">
    <el-card shadow="never">
      <div class="toolbar">
        <el-form inline @submit.prevent>
          <el-form-item label="关键字">
            <el-input
              v-model="query.keyword"
              placeholder="批次号 / 产品型号"
              clearable
              style="width: 220px"
              @keyup.enter="search"
              @clear="search"
            />
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="query.status" placeholder="全部" clearable style="width: 130px" @change="search">
              <el-option v-for="o in statusOptions" :key="o.value" :label="o.label" :value="o.value" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :icon="'Search'" @click="search">查询</el-button>
            <el-button :icon="'RefreshLeft'" @click="reset">重置</el-button>
          </el-form-item>
        </el-form>
        <el-button type="success" :icon="'Plus'" @click="createDialog?.open()">新增批次</el-button>
      </div>

      <el-table :data="list" v-loading="loading" stripe style="width: 100%">
        <el-table-column prop="lot_no" label="批次号" width="200">
          <template #default="{ row }">
            <el-link type="primary" :underline="false" @click="openDetail(row.id)">{{ row.lot_no }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="product" label="产品型号" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">{{ row.product || '—' }}</template>
        </el-table-column>
        <el-table-column prop="wafer_count" label="片数" width="80" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="LOT_STATUS_TAG[row.status as LotStatus]" size="small">
              {{ LOT_STATUS_LABELS[row.status as LotStatus] ?? row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="当前工艺" min-width="120">
          <template #default="{ row }">{{ row.current_process_name || '—' }}</template>
        </el-table-column>
        <el-table-column label="投产时间" width="160">
          <template #default="{ row }">{{ formatDateTime(row.received_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="90" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row.id)">加工记录</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无批次，点击右上角「新增批次」" />
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

    <LotCreateDialog ref="createDialog" @created="search" />
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
