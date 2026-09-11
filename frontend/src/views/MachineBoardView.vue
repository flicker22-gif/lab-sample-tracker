<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { useProductionStore } from '@/stores/production'
import { useMetaStore } from '@/stores/meta'
import { productionApi } from '@/api/production'
import { formatDateTime } from '@/utils/datetime'
import {
  MACHINE_STATUS_LABELS,
  MACHINE_STATUS_TAG,
} from '@/constants/production'
import type { Machine, MachineStatusValue, StatusLog } from '@/types/production'

const store = useProductionStore()
const meta = useMetaStore()

const loading = ref(false)

async function refresh() {
  loading.value = true
  try {
    await Promise.all([store.loadMeta(true), meta.load(true)])
  } finally {
    loading.value = false
  }
}
onMounted(refresh)

// 状态统计
const stats = computed(() => {
  const acc: Record<MachineStatusValue, number> = { idle: 0, running: 0, maintenance: 0, fault: 0 }
  for (const m of store.machines) acc[m.status]++
  return acc
})

// 按工艺顺序的分组
const groups = computed(() =>
  store.processes.map((p) => ({
    process: p,
    machines: store.machines.filter((m) => m.process_id === p.id),
  })),
)

// ---- 状态切换对话框 ----
const statusDialogVisible = ref(false)
const submitting = ref(false)
const statusFormRef = ref<FormInstance>()
const targetMachine = ref<Machine | null>(null)
const statusForm = reactive({
  status: 'maintenance' as Exclude<MachineStatusValue, 'running'>,
  operator_id: 0,
  reason: '',
})
const statusRules: FormRules = {
  status: [{ required: true, trigger: 'change' }],
  operator_id: [{ required: true, message: '请选择操作员', trigger: 'change' }],
}

const nextStatusOptions: { value: Exclude<MachineStatusValue, 'running'>; label: string; type: string }[] = [
  { value: 'idle', label: '恢复空闲', type: 'success' },
  { value: 'maintenance', label: '维护', type: 'info' },
  { value: 'fault', label: '故障', type: 'danger' },
]

function openStatusDialog(m: Machine, to: Exclude<MachineStatusValue, 'running'>) {
  targetMachine.value = m
  statusForm.status = to
  statusForm.operator_id = 0
  statusForm.reason = ''
  statusFormRef.value?.clearValidate()
  statusDialogVisible.value = true
}

async function submitStatus() {
  if (!statusFormRef.value || !targetMachine.value) return
  const valid = await statusFormRef.value.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    await productionApi.updateMachineStatus(targetMachine.value.id, {
      status: statusForm.status,
      operator_id: statusForm.operator_id,
      reason: statusForm.reason || undefined,
    })
    ElMessage.success('机台状态已更新')
    statusDialogVisible.value = false
    refresh()
  } finally {
    submitting.value = false
  }
}

// ---- 状态记录抽屉 ----
const drawerVisible = ref(false)
const logsLoading = ref(false)
const logsMachine = ref<Machine | null>(null)
const logs = ref<StatusLog[]>([])

async function openLogs(m: Machine) {
  logsMachine.value = m
  drawerVisible.value = true
  logsLoading.value = true
  try {
    logs.value = await productionApi.statusLogs(m.id)
  } finally {
    logsLoading.value = false
  }
}

const statusCards: { key: MachineStatusValue; label: string }[] = [
  { key: 'idle', label: '空闲' },
  { key: 'running', label: '加工中' },
  { key: 'maintenance', label: '维护' },
  { key: 'fault', label: '故障' },
]
</script>

<template>
  <div class="page" v-loading="loading">
    <el-card shadow="never" class="stat-card">
      <div class="stats">
        <div v-for="s in statusCards" :key="s.key" class="stat-item">
          <el-tag :type="MACHINE_STATUS_TAG[s.key]" size="large" effect="plain">{{ s.label }}</el-tag>
          <span class="stat-num">{{ stats[s.key] }}</span>
          <span class="stat-unit">台</span>
        </div>
        <el-button class="refresh-btn" :icon="'Refresh'" @click="refresh">刷新</el-button>
      </div>
    </el-card>

    <div v-for="g in groups" :key="g.process.id" class="process-block">
      <div class="process-title">
        <span class="process-name">{{ g.process.name }}</span>
        <el-tag size="small" effect="plain">{{ g.process.code }}</el-tag>
        <span class="muted">{{ g.machines.length }} 台机台</span>
      </div>
      <el-row :gutter="12">
        <el-col v-for="m in g.machines" :key="m.id" :xs="24" :sm="12" :md="8" :lg="6">
          <el-card shadow="hover" class="machine-card" :class="`st-${m.status}`">
            <div class="m-head">
              <span class="m-code">{{ m.code }}</span>
              <el-tag :type="MACHINE_STATUS_TAG[m.status]" size="small" effect="dark">
                {{ MACHINE_STATUS_LABELS[m.status] }}
              </el-tag>
            </div>
            <div class="m-name">{{ m.name }}</div>
            <div v-if="m.status === 'running' && m.current_lot_no" class="m-lot">
              <el-icon><Loading /></el-icon> 正在加工 <strong>{{ m.current_lot_no }}</strong>
            </div>
            <div v-else-if="m.remark" class="m-remark muted">{{ m.remark }}</div>
            <div class="m-actions">
              <template v-if="m.status === 'running'">
                <el-tooltip content="加工中的机台请到批次详情结束加工" placement="top">
                  <el-button size="small" disabled>占用中</el-button>
                </el-tooltip>
              </template>
              <template v-else>
                <el-button v-if="m.status !== 'idle'" size="small" type="success" plain
                  @click="openStatusDialog(m, 'idle')">恢复空闲</el-button>
                <el-button v-if="m.status !== 'maintenance'" size="small" plain
                  @click="openStatusDialog(m, 'maintenance')">维护</el-button>
                <el-button v-if="m.status !== 'fault'" size="small" type="danger" plain
                  @click="openStatusDialog(m, 'fault')">报故障</el-button>
              </template>
              <el-button size="small" link type="primary" @click="openLogs(m)">状态记录</el-button>
            </div>
          </el-card>
        </el-col>
        <el-col v-if="!g.machines.length" :span="24">
          <el-empty :description="`${g.process.name} 暂无机台`" :image-size="60" />
        </el-col>
      </el-row>
    </div>

    <!-- 状态切换 -->
    <el-dialog v-model="statusDialogVisible" title="机台状态登记" width="440px" destroy-on-close>
      <el-form ref="statusFormRef" :model="statusForm" :rules="statusRules" label-width="84px">
        <el-form-item label="机台">
          <el-input :model-value="`${targetMachine?.code} ${targetMachine?.name ?? ''}`" disabled />
        </el-form-item>
        <el-form-item label="变更为" prop="status">
          <el-radio-group v-model="statusForm.status">
            <el-radio-button v-for="o in nextStatusOptions" :key="o.value" :value="o.value">
              {{ o.label }}
            </el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="操作员" prop="operator_id">
          <el-select v-model="statusForm.operator_id" placeholder="请选择" style="width: 100%">
            <el-option v-for="u in meta.users" :key="u.id" :label="u.full_name" :value="u.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="原因/备注">
          <el-input v-model="statusForm.reason" type="textarea" :rows="2" placeholder="如：定期保养、设备异常说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="statusDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitStatus">确认</el-button>
      </template>
    </el-dialog>

    <!-- 状态变更记录 -->
    <el-drawer v-model="drawerVisible" size="420px" :title="`状态记录 · ${logsMachine?.code ?? ''}`">
      <div v-loading="logsLoading">
        <el-empty v-if="!logsLoading && !logs.length" description="暂无状态变更记录" />
        <el-timeline v-else>
          <el-timeline-item
            v-for="l in logs"
            :key="l.id"
            :type="MACHINE_STATUS_TAG[l.to_status] === 'danger' ? 'danger' : MACHINE_STATUS_TAG[l.to_status] === 'warning' ? 'warning' : 'primary'"
            :timestamp="formatDateTime(l.occurred_at)"
            placement="top"
          >
            <div class="log-line">
              <el-tag size="small" effect="plain">{{ MACHINE_STATUS_LABELS[(l.from_status || 'idle') as MachineStatusValue] || '—' }}</el-tag>
              <el-icon><Right /></el-icon>
              <el-tag size="small" :type="MACHINE_STATUS_TAG[l.to_status]" effect="dark">
                {{ MACHINE_STATUS_LABELS[l.to_status] }}
              </el-tag>
            </div>
            <div class="muted log-meta">{{ l.operator_name || '系统' }}<span v-if="l.reason"> · {{ l.reason }}</span></div>
          </el-timeline-item>
        </el-timeline>
      </div>
    </el-drawer>
  </div>
</template>

<style scoped>
.stat-card {
  margin-bottom: 16px;
}
.stats {
  display: flex;
  align-items: center;
  gap: 32px;
}
.stat-item {
  display: flex;
  align-items: center;
  gap: 8px;
}
.stat-num {
  font-size: 22px;
  font-weight: 700;
}
.stat-unit {
  color: #909399;
  font-size: 13px;
}
.refresh-btn {
  margin-left: auto;
}
.process-block {
  margin-bottom: 20px;
}
.process-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 4px 0 10px;
}
.process-name {
  font-size: 15px;
  font-weight: 600;
}
.muted {
  color: #909399;
  font-size: 12px;
}
.machine-card {
  margin-bottom: 12px;
  border-left: 3px solid #e4e7ed;
}
.machine-card.st-running {
  border-left-color: var(--el-color-warning);
}
.machine-card.st-fault {
  border-left-color: var(--el-color-danger);
}
.machine-card.st-maintenance {
  border-left-color: var(--el-color-info);
}
.machine-card.st-idle {
  border-left-color: var(--el-color-success);
}
.m-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.m-code {
  font-weight: 700;
  font-size: 15px;
}
.m-name {
  color: #606266;
  font-size: 13px;
  margin: 6px 0;
}
.m-lot {
  color: var(--el-color-warning);
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 4px;
}
.m-remark {
  margin-bottom: 6px;
}
.m-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 6px;
  border-top: 1px dashed #ebeef5;
  padding-top: 8px;
}
.log-line {
  display: flex;
  align-items: center;
  gap: 6px;
}
.log-meta {
  margin-top: 4px;
}
</style>
