<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { sampleApi } from '@/api'
import { useMetaStore } from '@/stores/meta'
import { ACTION_LABELS } from '@/constants/labels'
import type { CreateTransferPayload, Location, SampleDetail, TransferAction } from '@/types'

const props = defineProps<{ detail: SampleDetail | null }>()
const emit = defineEmits<{ (e: 'done'): void }>()

const meta = useMetaStore()
const visible = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()

const form = reactive<CreateTransferPayload>({
  from_location_id: null,
  to_location_id: 0,
  operator_id: 0,
  action: 'transfer',
  note: '',
})

const actions = (Object.keys(ACTION_LABELS) as TransferAction[])
  .filter((a) => a !== 'receive')
  .map((a) => ({ value: a, label: ACTION_LABELS[a] }))

const rules: FormRules = {
  to_location_id: [{ required: true, message: '请选择目标位置', trigger: 'change' }],
  operator_id: [{ required: true, message: '请选择操作人', trigger: 'change' }],
  action: [{ required: true, message: '请选择流转动作', trigger: 'change' }],
}

// 根据动作过滤可选目标位置
const targetLocations = computed<Location[]>(() => {
  switch (form.action) {
    case 'load':
      return meta.locations.filter((l) => l.type === 'device')
    case 'store':
      return meta.locations.filter((l) => l.type === 'fridge')
    case 'discard':
      return meta.locations.filter((l) => l.type === 'discard')
    default:
      return meta.locations
  }
})

// 切换动作时，若当前选择不在可选范围则清空
watch(
  () => form.action,
  () => {
    if (form.to_location_id && !targetLocations.value.some((l) => l.id === form.to_location_id)) {
      form.to_location_id = 0
    }
  },
)

async function open() {
  if (!props.detail) return
  await meta.load()
  form.from_location_id = props.detail.current_location_id
  form.to_location_id = 0
  form.operator_id = 0
  form.action = 'transfer'
  form.note = ''
  formRef.value?.clearValidate()
  visible.value = true
}

defineExpose({ open })

async function submit() {
  if (!props.detail || !formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  if (form.from_location_id && form.from_location_id === form.to_location_id) {
    ElMessage.warning('目标位置与来源位置相同')
    return
  }
  submitting.value = true
  try {
    await sampleApi.createTransfer(props.detail.id, {
      ...form,
      from_location_id: form.from_location_id ?? null,
    })
    ElMessage.success('流转记录已登记')
    visible.value = false
    emit('done')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <el-dialog v-model="visible" title="登记样品流转" width="520px" destroy-on-close>
    <el-form ref="formRef" :model="form" :rules="rules" label-width="92px">
      <el-form-item label="当前位置">
        <el-input :model-value="detail?.current_location_name || '—'" disabled />
      </el-form-item>
      <el-form-item label="流转动作" prop="action">
        <el-radio-group v-model="form.action">
          <el-radio-button v-for="a in actions" :key="a.value" :value="a.value">{{ a.label }}</el-radio-button>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="目标位置" prop="to_location_id">
        <el-select v-model="form.to_location_id" placeholder="移到哪台设备/哪个冰箱" style="width: 100%" filterable>
          <el-option
            v-for="l in targetLocations"
            :key="l.id"
            :label="l.type === 'fridge' ? `${l.name}（${l.temperature}℃ ${l.room}）` : `${l.name}（${l.room}）`"
            :value="l.id"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="操作人" prop="operator_id">
        <el-select v-model="form.operator_id" placeholder="谁操作的" style="width: 100%">
          <el-option v-for="u in meta.users" :key="u.id" :label="u.full_name" :value="u.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="form.note" type="textarea" :rows="2" placeholder="可选" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="submit">确认登记</el-button>
    </template>
  </el-dialog>
</template>
