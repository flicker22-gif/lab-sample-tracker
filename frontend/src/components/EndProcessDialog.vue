<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { productionApi } from '@/api/production'
import { useMetaStore } from '@/stores/meta'
import type { ProcessRecord } from '@/types/production'
import { formatDateTime } from '@/utils/datetime'
import { formatDuration } from '@/constants/production'

const props = defineProps<{ lotId: number | null; record: ProcessRecord | null }>()
const emit = defineEmits<{ (e: 'done'): void }>()

const meta = useMetaStore()
const visible = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()

// 打开对话框时按开始时间估算已加工时长（最终以后端结束时间为准）
const openedAt = ref(new Date())
const elapsedText = computed(() => {
  if (!props.record) return '—'
  const sec = Math.max(0, Math.floor((openedAt.value.getTime() - new Date(props.record.start_time).getTime()) / 1000))
  return formatDuration(sec)
})

const form = reactive({
  result: 'ok' as 'ok' | 'ng' | 'rework',
  wafer_out: 0,
  remark: '',
  operator_id: 0,
})

const rules: FormRules = {
  result: [{ required: true, message: '请选择加工结果', trigger: 'change' }],
}

async function open() {
  if (props.lotId == null || !props.record) return
  openedAt.value = new Date()
  await meta.load()
  form.result = 'ok'
  form.wafer_out = 0
  form.remark = ''
  form.operator_id = props.record.operator_id
  formRef.value?.clearValidate()
  visible.value = true
}
defineExpose({ open })

async function submit() {
  if (!formRef.value || props.lotId == null) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    await productionApi.end(props.lotId, {
      result: form.result,
      wafer_out: form.wafer_out,
      remark: form.remark || undefined,
      operator_id: form.operator_id || undefined,
    })
    ElMessage.success('加工已结束，机台恢复空闲')
    visible.value = false
    emit('done')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <el-dialog v-model="visible" title="结束加工" width="460px" destroy-on-close>
    <el-descriptions :column="1" border size="small" class="mb">
      <el-descriptions-item label="工艺/机台">
        {{ record?.process_name }} · {{ record?.machine_code }}
      </el-descriptions-item>
      <el-descriptions-item label="开始时间">
        {{ formatDateTime(record?.start_time) }}（已加工约 {{ elapsedText }}，最终以结束时间计算）
      </el-descriptions-item>
    </el-descriptions>

    <el-form ref="formRef" :model="form" :rules="rules" label-width="84px">
      <el-form-item label="加工结果" prop="result">
        <el-radio-group v-model="form.result">
          <el-radio-button value="ok">合格</el-radio-button>
          <el-radio-button value="ng">不合格</el-radio-button>
          <el-radio-button value="rework">返工</el-radio-button>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="产出片数">
        <el-input-number v-model="form.wafer_out" :min="0" :max="10000" />
      </el-form-item>
      <el-form-item label="结束操作人">
        <el-select v-model="form.operator_id" placeholder="默认开始操作员" clearable style="width: 100%">
          <el-option v-for="u in meta.users" :key="u.id" :label="u.full_name" :value="u.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="form.remark" type="textarea" :rows="2" placeholder="异常/处理说明" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="submit">确认结束</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.mb {
  margin-bottom: 16px;
}
</style>
