<template>
  <div>
    <el-card shadow="hover">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>其他服务列表</span>
          <el-button type="primary" @click="openForm('create')">新增服务</el-button>
        </div>
      </template>

      <el-form :inline="true" :model="search" style="margin-bottom: 10px">
        <el-form-item label="名称">
          <el-input v-model="search.keyword" placeholder="模糊搜索" clearable @clear="fetchData" @keyup.enter="fetchData" />
        </el-form-item>
        <el-form-item label="所属主机">
          <el-select v-model="search.machineId" placeholder="全部" clearable @change="fetchData" style="width: 160px" filterable>
            <el-option v-for="m in machineOptions" :key="m.id" :label="m.name" :value="m.id" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchData">查询</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="list" stripe border v-loading="loading" style="width: 100%">
        <el-table-column prop="name" label="服务名称" min-width="140" show-overflow-tooltip align="center" />
        <el-table-column prop="machineName" label="所属主机" width="140" show-overflow-tooltip align="center" />
        <el-table-column prop="machineIp" label="主机IP" width="150" align="center" />
        <el-table-column prop="port" label="端口" width="100" align="center" />
        <el-table-column prop="protocol" label="协议" width="70" align="center" />
        <el-table-column prop="egressCount" label="出站" width="65" align="center" />
        <el-table-column prop="remark" label="备注" min-width="160" show-overflow-tooltip align="center" />
        <el-table-column label="操作" width="220" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="info" link size="small" @click="viewDetail(row)">查看</el-button>
            <el-button type="primary" link size="small" @click="openForm('edit', row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        style="margin-top: 16px; justify-content: flex-end"
        @change="fetchData"
      />
    </el-card>

    <el-dialog v-model="formVisible" :title="formMode === 'create' ? '新增服务' : '编辑服务'" width="480px" :close-on-click-modal="false" :lock-scroll="false">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="所属主机" prop="machineId">
          <el-select v-model="form.machineId" placeholder="请选择" style="width: 100%" filterable :teleported="false">
            <el-option v-for="m in machineOptions" :key="m.id" :label="m.name" :value="m.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="服务名称" prop="name">
          <el-input v-model="form.name" placeholder="例如：Nginx" />
        </el-form-item>
        <el-form-item label="端口" prop="port">
          <el-input-number v-model="form.port" :min="1" :max="65535" style="width: 100%" />
        </el-form-item>
        <el-form-item label="协议">
          <el-select v-model="form.protocol" style="width: 100%">
            <el-option label="TCP" value="TCP" />
            <el-option label="UDP" value="UDP" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="authStore.isAdmin" label="公共服务">
          <el-switch v-model="form.isPublic" active-text="是" inactive-text="否" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" title="服务详情" width="480px" :close-on-click-modal="false" :lock-scroll="false">
      <template v-if="detailData">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="服务名称">{{ detailData.name }}</el-descriptions-item>
          <el-descriptions-item label="所属主机">{{ detailData.machineName }}</el-descriptions-item>
          <el-descriptions-item label="主机IP">{{ detailData.machineIp || '-' }}</el-descriptions-item>
          <el-descriptions-item label="端口">{{ detailData.port || '-' }}</el-descriptions-item>
          <el-descriptions-item label="协议">{{ detailData.protocol || 'TCP' }}</el-descriptions-item>
          <el-descriptions-item label="出站数量">{{ detailData.egressCount || 0 }}</el-descriptions-item>
          <el-descriptions-item label="备注">{{ detailData.remark || '-' }}</el-descriptions-item>
        </el-descriptions>
      </template>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { getOtherServices, createOtherService, updateOtherService, deleteOtherService } from '../api/otherService'
import { getMachines } from '../api/machine'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()

const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const machineOptions = ref([])
const search = reactive({ keyword: '', machineId: '' })
const formVisible = ref(false)
const formMode = ref('create')
const formRef = ref(null)
const submitting = ref(false)
const editId = ref(null)

const detailVisible = ref(false)
const detailData = ref(null)

const form = reactive({
  machineId: null, name: '', port: 80, protocol: 'TCP', isPublic: false, remark: ''
})

const rules = {
  machineId: [{ required: true, message: '请选择所属主机', trigger: 'change' }],
  name: [{ required: true, message: '请输入服务名称', trigger: 'blur' }],
  port: [{ required: true, message: '请输入端口号', trigger: 'blur' }]
}

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getOtherServices({ page: page.value, pageSize: pageSize.value, ...search })
    list.value = res.data.list
    total.value = res.data.total
  } catch {
  } finally {
    loading.value = false
  }
}

const fetchMachines = async () => {
  try {
    const res = await getMachines({ page: 1, pageSize: 200 })
    machineOptions.value = res.data.list
  } catch {}
}

const openForm = (mode, row) => {
  formMode.value = mode
  formVisible.value = true
  if (mode === 'edit' && row) {
    editId.value = row.id
    Object.assign(form, {
      machineId: row.machineId, name: row.name,
      port: row.port, protocol: row.protocol || 'TCP',
      isPublic: !!row.isPublic, remark: row.remark || ''
    })
  } else {
    editId.value = null
    Object.assign(form, { machineId: null, name: '', port: 80, protocol: 'TCP', isPublic: false, remark: '' })
  }
}

const handleSubmit = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    if (formMode.value === 'create') {
      await createOtherService(form)
      ElMessage.success('创建成功')
    } else {
      await updateOtherService(editId.value, form)
      ElMessage.success('更新成功')
    }
    formVisible.value = false
    fetchData()
  } catch {
  } finally {
    submitting.value = false
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(`确定删除服务「${row.name}」吗？`, '确认删除', {
    confirmButtonText: '确定删除',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    await deleteOtherService(row.id)
    ElMessage.success('删除成功')
    fetchData()
  }).catch(() => {})
}

const viewDetail = (row) => {
  detailData.value = row
  detailVisible.value = true
}

onMounted(() => {
  fetchData()
  fetchMachines()
})
</script>