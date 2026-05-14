<template>
  <div>
    <el-card shadow="hover">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>出站方式列表</span>
          <el-button type="primary" @click="openForm('create')">新增出站方式</el-button>
        </div>
      </template>

      <el-form :inline="true" :model="search" style="margin-bottom: 10px">
        <el-form-item label="所属服务">
          <el-select v-model="search.serviceId" placeholder="全部" clearable @change="fetchData" style="width: 160px" filterable>
            <el-option v-for="s in serviceOptions" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="方式类型">
          <el-select v-model="search.methodType" placeholder="全部" clearable @change="fetchData" style="width: 140px">
            <el-option label="FRP 内网穿透" value="FRP" />
            <el-option label="端口映射" value="PORT_MAPPING" />
            <el-option label="VPN" value="VPN" />
            <el-option label="直接访问" value="DIRECT" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="search.status" placeholder="全部" clearable @change="fetchData" style="width: 120px">
            <el-option label="启用" :value="1" />
            <el-option label="停用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchData">查询</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="list" stripe border v-loading="loading">
        <el-table-column prop="serviceName" label="所属服务" min-width="130" />
        <el-table-column prop="machineName" label="所属主机" min-width="130" />
        <el-table-column label="方式类型" width="120">
          <template #default="{ row }">
            <el-tag :type="methodTypeTag(row.methodType)" size="small">{{ methodTypeLabel(row.methodType) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="proxyName" label="代理名称" min-width="120" />
        <el-table-column label="公网地址" min-width="180">
          <template #default="{ row }">
            {{ row.publicIp }}:{{ row.publicPort }}
          </template>
        </el-table-column>
        <el-table-column label="内网地址" min-width="180">
          <template #default="{ row }">
            {{ row.internalIp }}:{{ row.internalPort }}
          </template>
        </el-table-column>
        <el-table-column prop="protocol" label="协议" width="70" align="center" />
        <el-table-column label="状态" width="70" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'warning'" size="small">
              {{ row.status === 1 ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="success" link size="small" @click="copyAddress(row)">复制地址</el-button>
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

    <el-dialog v-model="formVisible" :title="formMode === 'create' ? '新增出站方式' : '编辑出站方式'" width="560px" :close-on-click-modal="false" :lock-scroll="false">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item label="所属服务" prop="serviceId">
          <el-select v-model="form.serviceId" placeholder="请选择" style="width: 100%" filterable :teleported="false">
            <el-option v-for="s in serviceOptions" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="方式类型" prop="methodType">
          <el-select v-model="form.methodType" placeholder="请选择" style="width: 100%">
            <el-option label="FRP 内网穿透" value="FRP" />
            <el-option label="端口映射" value="PORT_MAPPING" />
            <el-option label="VPN" value="VPN" />
            <el-option label="直接访问" value="DIRECT" />
          </el-select>
        </el-form-item>
        <el-form-item label="代理/隧道名称">
          <el-input v-model="form.proxyName" placeholder="例如：nginx-frp-tunnel" />
        </el-form-item>
        <el-form-item label="公网IP" prop="publicIp">
          <el-input v-model="form.publicIp" placeholder="对外公网IP" />
        </el-form-item>
        <el-form-item label="公网端口" prop="publicPort">
          <el-input-number v-model="form.publicPort" :min="1" :max="65535" style="width: 100%" />
        </el-form-item>
        <el-form-item label="内网IP" prop="internalIp">
          <el-input v-model="form.internalIp" placeholder="内网IP地址" />
        </el-form-item>
        <el-form-item label="内网端口" prop="internalPort">
          <el-input-number v-model="form.internalPort" :min="1" :max="65535" style="width: 100%" />
        </el-form-item>
        <el-form-item label="协议">
          <el-select v-model="form.protocol" style="width: 100%">
            <el-option label="TCP" value="TCP" />
            <el-option label="UDP" value="UDP" />
            <el-option label="HTTP" value="HTTP" />
            <el-option label="HTTPS" value="HTTPS" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="停用" />
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
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { getEgressMethods, createEgressMethod, updateEgressMethod, deleteEgressMethod } from '../api/egress'
import { getServices } from '../api/service'
import { ElMessage, ElMessageBox } from 'element-plus'

const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const loading = ref(false)
const serviceOptions = ref([])
const search = reactive({ serviceId: '', methodType: '', status: '' })
const formVisible = ref(false)
const formMode = ref('create')
const formRef = ref(null)
const submitting = ref(false)
const editId = ref(null)

const form = reactive({
  serviceId: null, methodType: 'FRP', proxyName: '',
  publicIp: '', publicPort: 8080, internalIp: '', internalPort: 80,
  protocol: 'TCP', status: 1, remark: ''
})

const rules = {
  serviceId: [{ required: true, message: '请选择所属服务', trigger: 'change' }],
  methodType: [{ required: true, message: '请选择方式类型', trigger: 'change' }],
  publicIp: [{ required: true, message: '请输入公网IP', trigger: 'blur' }],
  publicPort: [{ required: true, message: '请输入公网端口', trigger: 'blur' }],
  internalIp: [{ required: true, message: '请输入内网IP', trigger: 'blur' }],
  internalPort: [{ required: true, message: '请输入内网端口', trigger: 'blur' }]
}

const methodTypeLabel = (type) => {
  const map = { FRP: 'FRP内网穿透', PORT_MAPPING: '端口映射', VPN: 'VPN', DIRECT: '直接访问' }
  return map[type] || type
}

const methodTypeTag = (type) => {
  const map = { FRP: 'success', PORT_MAPPING: '', VPN: 'warning', DIRECT: 'info' }
  return map[type] || 'info'
}

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getEgressMethods({ page: page.value, pageSize: pageSize.value, ...search })
    list.value = res.data.list
    total.value = res.data.total
  } catch {
  } finally {
    loading.value = false
  }
}

const fetchServices = async () => {
  try {
    const res = await getServices({ page: 1, pageSize: 200 })
    serviceOptions.value = res.data.list
  } catch {}
}

const openForm = (mode, row) => {
  formMode.value = mode
  formVisible.value = true
  if (mode === 'edit' && row) {
    editId.value = row.id
    Object.assign(form, {
      serviceId: row.serviceId, methodType: row.methodType, proxyName: row.proxyName || '',
      publicIp: row.publicIp, publicPort: row.publicPort,
      internalIp: row.internalIp, internalPort: row.internalPort,
      protocol: row.protocol || 'TCP', status: row.status, remark: row.remark || ''
    })
  } else {
    editId.value = null
    Object.assign(form, {
      serviceId: '', methodType: 'FRP', proxyName: '',
      publicIp: '', publicPort: 8080, internalIp: '', internalPort: 80,
      protocol: 'TCP', status: 1, remark: ''
    })
  }
}

const handleSubmit = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    if (formMode.value === 'create') {
      await createEgressMethod(form)
      ElMessage.success('创建成功')
    } else {
      await updateEgressMethod(editId.value, form)
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
  ElMessageBox.confirm(`确定删除该出站方式吗？`, '确认删除', {
    confirmButtonText: '确定删除',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    await deleteEgressMethod(row.id)
    ElMessage.success('删除成功')
    fetchData()
  }).catch(() => {})
}

const copyAddress = (row) => {
  const addr = `${row.protocol.toLowerCase()}://${row.publicIp}:${row.publicPort}`
  navigator.clipboard.writeText(addr).then(() => {
    ElMessage.success(`已复制: ${addr}`)
  }).catch(() => {
    ElMessage.warning('复制失败，请手动复制')
  })
}

onMounted(() => {
  fetchData()
  fetchServices()
})
</script>