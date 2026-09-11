<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { productionApi } from '@/api/production'
import { useMetaStore } from '@/stores/meta'
import { useProductionStore } from '@/stores/production'
import { MACHINE_STATUS_LABELS, MACHINE_STATUS_TAG } from '@/constants/production'
import type { MachineStatusValue } from '@/types/production'

const props = defineProps<{ lotId: number | null }>()
const emit = defineEmits<{ (e: 'done'): void }>()

const meta = useMetaStore()
const prod = useProductionStore()

const visible = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()

const form = reactive({
  process_id: 0,
  machine_id: 0,
  operator_id: 0,
  remark: '',
})

const rules: FormRules = {
  process_id: [{ required: true, message: '请选择工艺', trigger: 'change' }],
  machine_id: [{ required: true, message: '请选择机台', trigger: 'change' }],
  operator_id: [{ required: true, message: '请选择操作员', trigger: 'change' }],
}

// 当前工艺下的机台
const machinesInProcess = computed(() =>
  form.process_id ? prod.machines.filter((m) => m.process_id === form.process_id) : [],
)

const selectedMachine = computed(() => prod.machines.find((m) => m.id === form.machine_id))

async function open() {
  if (props.lotId == null) return
  await Promise.all([meta.load(), prod.loadMeta()])
  form.process_id = prod.processes[0]?.id ?? 0
  form.machine_id = 0
  form.operator_id = 0
  form.remark = ''
  formRef.value?.clearValidate()
  visible.value = true
}
defineExpose({ open })

// 切工艺时清空机台选择
watch(
  () => form.process_id,
  () => {
    form.machine_id = 0
  },
)

function machineTagType(s: MachineStatusValue) {
  return MACHINE_STATUS_TAG[s]
}

async function submit() {
  if (!formRef.value || props.lotId == null) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  const machine = selectedMachine.value
  if (!machine || machine.status !== 'idle') {
    ElMessage.warning('只能选择空闲机台开始加工')
    return
  }
  submitting.value = true
  try {
    await productionApi.start(props.lotId, {
      machine_id: form.machine_id,
      operator_id: form.operator_id,
      remark: form.remark || undefined,
    })
    ElMessage.success(`已在 ${machine.code} 开始加工`)
    visible.value = false
    emit('done')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <el-dialog v-model="visible" title="开始加工（选择机台）" width="560px" destroy-on-close>
    <el-form ref="formRef" :model="form" :rules="rules" label-width="84px">
      <el-form-item label="工艺" prop="process_id">
        <el-select v-model="form.process_id" placeholder="选择工艺" style="width: 100%">
          <el-option
            v-for="p in prod.processes"
            :key="p.id"
            :label="`${p.name}（${p.code}）`"
            :value="p.id"
          />
        </el-select>
      </el-form-item>

      <el-form-item label="机台" prop="machine_id">
        <el-radio-group v-model="form.machine_id" class="machine-group">
          <el-radio
            v-for="m in machinesInProcess"
            :key="m.id"
            :value="m.id"
            :disabled="m.status !== 'idle'"
            class="machine-radio"
          >
            <div class="machine-card">
              <div class="machine-card-head">
                <span class="machine-code">{{ m.code }}</span>
                <el-tag :type="machineTagType(m.status)" size="small">
                  {{ MACHINE_STATUS_LABELS[m.status] }}
                </el-tag>
              </div>
              <div class="machine-name">{{ m.name }}</div>
              <div v-if="m.status === 'running'" class="machine-sub">
                正在加工 {{ m.current_lot_no }}
              </div>
              <div v-else-if="m.remark" class="machine-sub">{{ m.remark }}</div>
            </div>
          </el-radio>
          <el-empty v-if="form.process_id && !machinesInProcess.length" description="该工艺暂无机台" :image-size="60" />
        </el-radio-group>
      </el-form-item>

      <el-form-item label="操作员" prop="operator_id">
        <el-select v-model="form.operator_id" placeholder="选择操作员" style="width: 100%">
          <el-option v-for="u in meta.users" :key="u.id" :label="u.full_name" :value="u.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="form.remark" type="textarea" :rows="2" placeholder="可选" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="submit">开始加工</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.machine-group {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  width: 100%;
}
.machine-radio {
  margin: 0;
  height: auto;
  align-items: flex-start;
  padding: 8px 10px;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  width: calc(50% - 4px);
  box-sizing: border-box;
}
.machine-radio.is-checked {
  border-color: var(--el-color-primary);
}
.machine-radio.is-disabled {
  background: #f5f7fa;
}
.machine-card {
  margin-left: 6px;
}
.machine-card-head {
  display: flex;
  align-items: center;
  gap: 8px;
}
.machine-code {
  font-weight: 600;
}
.machine-name {
  color: #606266;
  font-size: 13px;
  margin-top: 2px;
}
.machine-sub {
  color: #909399;
  font-size: 12px;
  margin-top: 2px;
}
</style>
