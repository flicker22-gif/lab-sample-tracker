<script setup lang="ts">
import { reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { sampleApi } from '@/api'
import { useMetaStore } from '@/stores/meta'
import type { CreateResultPayload } from '@/types'

const props = defineProps<{ sampleId: number | null }>()
const emit = defineEmits<{ (e: 'done'): void }>()

const meta = useMetaStore()
const visible = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()

const initial = (): CreateResultPayload => ({
  item_name: '',
  method: '',
  instrument: '',
  result_value: '',
  unit: '',
  conclusion: '',
  report_no: '',
  analyst_id: 0,
})
const form = reactive<CreateResultPayload>(initial())

const rules: FormRules = {
  item_name: [{ required: true, message: '请输入检测项目', trigger: 'blur' }],
  analyst_id: [{ required: true, message: '请选择检测人员', trigger: 'change' }],
}

const conclusionOptions = [
  { value: 'na', label: '待定（仅登记项目）' },
  { value: 'qualified', label: '合格' },
  { value: 'unqualified', label: '不合格' },
]

async function open() {
  if (props.sampleId == null) return
  Object.assign(form, initial())
  await meta.load()
  formRef.value?.clearValidate()
  visible.value = true
}

defineExpose({ open })

async function submit() {
  if (!formRef.value || props.sampleId == null) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    await sampleApi.createResult(props.sampleId, { ...form })
    ElMessage.success('检测结果已登记')
    visible.value = false
    emit('done')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <el-dialog v-model="visible" title="登记检测项目 / 结果" width="540px" destroy-on-close>
    <el-form ref="formRef" :model="form" :rules="rules" label-width="96px">
      <el-form-item label="检测项目" prop="item_name">
        <el-input v-model="form.item_name" placeholder="如：铅(Pb)、菌落总数" maxlength="128" />
      </el-form-item>
      <div class="two-col">
        <el-form-item label="检测方法">
          <el-input v-model="form.method" placeholder="如 GB 5749-2022" maxlength="128" />
        </el-form-item>
        <el-form-item label="使用仪器">
          <el-input v-model="form.instrument" placeholder="如 ICP-MS-01" maxlength="128" />
        </el-form-item>
      </div>
      <div class="two-col">
        <el-form-item label="结果值">
          <el-input v-model="form.result_value" placeholder="如 0.003" />
        </el-form-item>
        <el-form-item label="单位">
          <el-input v-model="form.unit" placeholder="如 mg/L" maxlength="32" />
        </el-form-item>
      </div>
      <el-form-item label="结论">
        <el-select v-model="form.conclusion" placeholder="选择判定结论" style="width: 100%" clearable>
          <el-option v-for="o in conclusionOptions" :key="o.value" :label="o.label" :value="o.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="报告编号">
        <el-input v-model="form.report_no" placeholder="填写报告编号后样品自动标记为「已出报告」" maxlength="64" />
      </el-form-item>
      <el-form-item label="检测人员" prop="analyst_id">
        <el-select v-model="form.analyst_id" placeholder="请选择" style="width: 100%">
          <el-option v-for="u in meta.users" :key="u.id" :label="u.full_name" :value="u.id" />
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="submit">保存</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.two-col {
  display: flex;
  gap: 0 16px;
}
.two-col .el-form-item {
  flex: 1;
}
</style>
