<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useSampleStore } from '@/stores/sample'
import { formatDateTime } from '@/utils/datetime'
import {
  ACTION_LABELS,
  ACTION_TIMELINE_TYPE,
  CONCLUSION_LABELS,
  STATUS_LABELS,
  STATUS_TAG_TYPE,
} from '@/constants/labels'
import type { SampleStatus, Transfer } from '@/types'
import TransferDialog from '@/components/TransferDialog.vue'
import ResultDialog from '@/components/ResultDialog.vue'

const props = defineProps<{ id: number }>()

const router = useRouter()
const store = useSampleStore()
const { detail, detailLoading } = storeToRefs(store)

const transferDialog = ref<InstanceType<typeof TransferDialog>>()
const resultDialog = ref<InstanceType<typeof ResultDialog>>()

async function load() {
  await store.fetchDetail(props.id)
}

watch(
  () => props.id,
  () => load(),
  { immediate: true },
)

// 时间线按时间倒序展示，最新动态在最上面
const timeline = computed<Transfer[]>(() =>
  detail.value ? [...detail.value.transfers].reverse() : [],
)

function goBack() {
  router.push({ name: 'samples' })
}

function locTagType(type: string): 'primary' | 'success' | 'info' | 'warning' | 'danger' {
  if (type === 'fridge') return 'primary'
  if (type === 'device') return 'warning'
  if (type === 'discard') return 'danger'
  return 'info'
}
</script>

<template>
  <div class="page" v-loading="detailLoading">
    <div class="page-head">
      <el-button :icon="'ArrowLeft'" link @click="goBack">返回列表</el-button>
      <el-space>
        <el-button type="primary" plain :icon="'Position'" @click="transferDialog?.open()">
          登记流转
        </el-button>
        <el-button type="success" plain :icon="'Document'" @click="resultDialog?.open()">
          登记检测结果
        </el-button>
      </el-space>
    </div>

    <template v-if="detail">
      <!-- 样品基本信息 -->
      <el-card shadow="never" class="info-card">
        <template #header>
          <div class="card-header">
            <div class="title-line">
              <span class="sample-code">{{ detail.code }}</span>
              <el-tag :type="STATUS_TAG_TYPE[detail.status as SampleStatus]">
                {{ STATUS_LABELS[detail.status as SampleStatus] ?? detail.status }}
              </el-tag>
            </div>
            <span class="sample-name">{{ detail.name }}</span>
          </div>
        </template>
        <el-descriptions :column="3" border size="default">
          <el-descriptions-item label="样品类型">{{ detail.category || '—' }}</el-descriptions-item>
          <el-descriptions-item label="送检单位">{{ detail.source || '—' }}</el-descriptions-item>
          <el-descriptions-item label="收样人">{{ detail.receiver_name }}</el-descriptions-item>
          <el-descriptions-item label="收样时间">{{ formatDateTime(detail.received_at) }}</el-descriptions-item>
          <el-descriptions-item label="当前位置">
            <el-tag v-if="detail.current_location_name" :type="locTagType(detail.current_location_type)" size="small">
              {{ detail.current_location_name }}
            </el-tag>
            <span v-else>—</span>
          </el-descriptions-item>
          <el-descriptions-item label="备注">{{ detail.remark || '—' }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <el-row :gutter="16" class="body-row">
        <!-- 流转轨迹 -->
        <el-col :xs="24" :md="14">
          <el-card shadow="never" class="timeline-card">
            <template #header>
              <div class="card-header">
                <span class="block-title"><el-icon><Position /></el-icon> 流转轨迹</span>
                <span class="muted">共 {{ detail.transfers.length }} 条记录</span>
              </div>
            </template>

            <el-empty v-if="!timeline.length" description="暂无流转记录" :image-size="80" />
            <el-timeline v-else class="track-timeline">
              <el-timeline-item
                v-for="t in timeline"
                :key="t.id"
                :type="ACTION_TIMELINE_TYPE[t.action]"
                :timestamp="formatDateTime(t.occurred_at)"
                placement="top"
                size="large"
              >
                <div class="track-node">
                  <div class="track-title">
                    <el-tag :type="ACTION_TIMELINE_TYPE[t.action]" size="small" effect="dark">
                      {{ ACTION_LABELS[t.action] ?? t.action }}
                    </el-tag>
                    <span class="operator">{{ t.operator_name }}</span>
                  </div>
                  <div class="track-route">
                    <el-tag v-if="t.from_location_name" :type="locTagType(t.from_location_type)" size="small" effect="plain">
                      {{ t.from_location_name }}
                    </el-tag>
                    <span v-else class="muted">起点（收样）</span>
                    <el-icon class="arrow"><Right /></el-icon>
                    <el-tag :type="locTagType(t.to_location_type)" size="small">
                      {{ t.to_location_name }}
                    </el-tag>
                  </div>
                  <div v-if="t.note" class="track-note">备注：{{ t.note }}</div>
                </div>
              </el-timeline-item>
            </el-timeline>
          </el-card>
        </el-col>

        <!-- 检测项目与结果 -->
        <el-col :xs="24" :md="10">
          <el-card shadow="never">
            <template #header>
              <div class="card-header">
                <span class="block-title"><el-icon><DataAnalysis /></el-icon> 检测项目与结果</span>
                <span class="muted">{{ detail.results.length }} 项</span>
              </div>
            </template>
            <el-empty v-if="!detail.results.length" description="尚未登记检测结果" :image-size="80" />
            <el-table v-else :data="detail.results" size="small" border>
              <el-table-column label="项目 / 方法" min-width="150">
                <template #default="{ row }">
                  <div class="result-item">{{ row.item_name }}</div>
                  <div class="result-sub muted">{{ row.method || '—' }}</div>
                </template>
              </el-table-column>
              <el-table-column label="结果" min-width="110">
                <template #default="{ row }">
                  <div>{{ row.result_value || '—' }} {{ row.unit }}</div>
                  <div class="result-sub muted">{{ row.instrument }}</div>
                </template>
              </el-table-column>
              <el-table-column label="结论 / 报告" width="110">
                <template #default="{ row }">
                  <el-tag
                    v-if="row.conclusion"
                    size="small"
                    :type="row.conclusion === 'unqualified' ? 'danger' : row.conclusion === 'qualified' ? 'success' : 'info'"
                  >
                    {{ CONCLUSION_LABELS[row.conclusion] ?? row.conclusion }}
                  </el-tag>
                  <div v-else class="muted">—</div>
                  <div v-if="row.report_no" class="result-sub">{{ row.report_no }}</div>
                </template>
              </el-table-column>
              <el-table-column label="检测人/时间" width="110">
                <template #default="{ row }">
                  <div>{{ row.analyst_name }}</div>
                  <div class="result-sub muted">{{ formatDateTime(row.tested_at) }}</div>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-col>
      </el-row>
    </template>

    <el-empty v-else-if="!detailLoading" description="样品不存在或已被删除">
      <el-button type="primary" @click="goBack">返回列表</el-button>
    </el-empty>

    <TransferDialog ref="transferDialog" :detail="detail" @done="load" />
    <ResultDialog ref="resultDialog" :sample-id="detail?.id ?? null" @done="load" />
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
.sample-code {
  font-size: 18px;
  font-weight: 700;
  color: #303133;
  letter-spacing: 0.5px;
}
.sample-name {
  color: #606266;
  font-size: 13px;
}
.block-title {
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.muted {
  color: #909399;
  font-size: 12px;
}
.body-row {
  margin: 0 !important;
}
.track-timeline {
  padding: 8px 4px 0 8px;
  max-height: 640px;
  overflow-y: auto;
}
.track-node {
  padding-bottom: 4px;
}
.track-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}
.operator {
  color: #606266;
  font-size: 13px;
}
.track-route {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.arrow {
  color: #c0c4cc;
}
.track-note {
  margin-top: 4px;
  color: #909399;
  font-size: 12px;
}
.result-item {
  font-weight: 600;
  color: #303133;
}
.result-sub {
  font-size: 12px;
  margin-top: 2px;
}
</style>
