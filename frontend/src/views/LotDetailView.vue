<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useProductionStore } from '@/stores/production'
import { formatDateTime } from '@/utils/datetime'
import { formatDuration, LOT_STATUS_LABELS, LOT_STATUS_TAG, RESULT_LABELS, RESULT_TAG } from '@/constants/production'
import type { LotStatus, ProcessRecord, ProcessResultValue } from '@/types/production'
import { productionApi } from '@/api/production'
import StartProcessDialog from '@/components/StartProcessDialog.vue'
import EndProcessDialog from '@/components/EndProcessDialog.vue'

const props = defineProps<{ id: number }>()
const router = useRouter()
const store = useProductionStore()
const { detail, detailLoading } = storeToRefs(store)

const startDialog = ref<InstanceType<typeof StartProcessDialog>>()
const endDialog = ref<InstanceType<typeof EndProcessDialog>>()
const endingRecord = ref<ProcessRecord | null>(null)
const completing = ref(false)

async function load() {
  await Promise.all([store.loadMeta(), store.fetchDetail(props.id)])
}
watch(() => props.id, load, { immediate: true })

// 进行中的加工记录（end_time 为空）
const openRecord = computed(() => detail.value?.records.find((r) => !r.end_time) ?? null)
const canStart = computed(
  () => detail.value && detail.value.status !== 'completed' && detail.value.status !== 'scrapped' && !openRecord.value,
)
const canComplete = computed(
  () => detail.value && detail.value.status !== 'completed' && detail.value.status !== 'scrapped' && !openRecord.value,
)

function onEnd() {
  if (!openRecord.value) return
  endingRecord.value = openRecord.value
  endDialog.value?.open()
}

async function completeLot() {
  await ElMessageBox.confirm('确认该批次全部工艺加工完成并标记为「已完工」？', '批次完工', {
    type: 'warning',
  }).catch(() => {
    throw new Error('cancel')
  })
  completing.value = true
  try {
    await productionApi.complete(props.id)
    ElMessage.success('批次已完工')
    load()
  } finally {
    completing.value = false
  }
}

function goBack() {
  router.push({ name: 'lots' })
}
</script>

<template>
  <div class="page" v-loading="detailLoading">
    <div class="page-head">
      <el-button :icon="'ArrowLeft'" link @click="goBack">返回批次列表</el-button>
      <el-space>
        <el-button v-if="openRecord" type="warning" plain :icon="'VideoPause'" @click="onEnd">结束加工</el-button>
        <el-button v-if="canStart" type="primary" :icon="'VideoPlay'" @click="startDialog?.open()">开始加工</el-button>
        <el-button v-if="canComplete" type="success" plain :icon="'CircleCheck'" :loading="completing" @click="completeLot">
          批次完工
        </el-button>
      </el-space>
    </div>

    <template v-if="detail">
      <el-card shadow="never" class="info-card">
        <template #header>
          <div class="card-header">
            <div class="title-line">
              <span class="lot-no">{{ detail.lot_no }}</span>
              <el-tag :type="LOT_STATUS_TAG[detail.status as LotStatus]">
                {{ LOT_STATUS_LABELS[detail.status as LotStatus] ?? detail.status }}
              </el-tag>
              <el-tag v-if="openRecord" type="warning" effect="dark">
                正在 {{ openRecord.process_name }} · {{ openRecord.machine_code }}
              </el-tag>
            </div>
            <span class="lot-sub">{{ detail.product || '—' }} · {{ detail.wafer_count }} 片</span>
          </div>
        </template>
        <el-descriptions :column="3" border size="default">
          <el-descriptions-item label="产品型号">{{ detail.product || '—' }}</el-descriptions-item>
          <el-descriptions-item label="晶圆数量">{{ detail.wafer_count }} 片</el-descriptions-item>
          <el-descriptions-item label="当前工艺">{{ detail.current_process_name || '—' }}</el-descriptions-item>
          <el-descriptions-item label="投产时间">{{ formatDateTime(detail.received_at) }}</el-descriptions-item>
          <el-descriptions-item label="备注" :span="2">{{ detail.remark || '—' }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <el-card shadow="never">
        <template #header>
          <span class="block-title"><el-icon><Histogram /></el-icon> 加工记录（按工艺顺序追溯）</span>
        </template>
        <el-table :data="detail.records" border>
          <el-table-column type="index" label="#" width="50" />
          <el-table-column label="工艺" min-width="130">
            <template #default="{ row }">
              <div class="strong">{{ row.process_name }}</div>
              <div class="muted">{{ row.process_code }}</div>
            </template>
          </el-table-column>
          <el-table-column label="机台" min-width="130">
            <template #default="{ row }">
              <div class="strong">{{ row.machine_code }}</div>
              <div class="muted">{{ row.machine_name }}</div>
            </template>
          </el-table-column>
          <el-table-column prop="operator_name" label="操作员" width="90" />
          <el-table-column label="开始时间" width="160">
            <template #default="{ row }">{{ formatDateTime(row.start_time) }}</template>
          </el-table-column>
          <el-table-column label="结束时间" width="160">
            <template #default="{ row }">
              <span v-if="row.end_time">{{ formatDateTime(row.end_time) }}</span>
              <el-tag v-else type="warning" size="small">加工中</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="时长" width="100">
            <template #default="{ row }">{{ formatDuration(row.duration_sec) }}</template>
          </el-table-column>
          <el-table-column label="结果/产出" width="120">
            <template #default="{ row }">
              <el-tag
                v-if="row.result !== 'pending'"
                :type="RESULT_TAG[row.result as ProcessResultValue]"
                size="small"
              >
                {{ RESULT_LABELS[row.result as ProcessResultValue] }}
              </el-tag>
              <span v-else class="muted">—</span>
              <div v-if="row.wafer_out" class="muted">{{ row.wafer_out }} 片</div>
            </template>
          </el-table-column>
          <el-table-column prop="remark" label="备注" min-width="120" show-overflow-tooltip>
            <template #default="{ row }">{{ row.remark || '—' }}</template>
          </el-table-column>
          <template #empty>
            <el-empty description="还没有加工记录，点击右上角「开始加工」" />
          </template>
        </el-table>
      </el-card>
    </template>

    <StartProcessDialog ref="startDialog" :lot-id="detail?.id ?? null" @done="load" />
    <EndProcessDialog ref="endDialog" :lot-id="detail?.id ?? null" :record="endingRecord" @done="load" />
  </div>
</template>

<style scoped>
.page-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.info-card {
  margin-bottom: 16px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}
.title-line {
  display: flex;
  align-items: center;
  gap: 10px;
}
.lot-no {
  font-size: 18px;
  font-weight: 700;
}
.lot-sub {
  color: #606266;
  font-size: 13px;
}
.block-title {
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.strong {
  font-weight: 600;
}
.muted {
  color: #909399;
  font-size: 12px;
  margin-top: 2px;
}
</style>
