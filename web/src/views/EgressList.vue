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
            <el-option v-for="s in allServiceOptions" :key="s.value" :label="s.label" :value="s.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="出站服务">
          <el-select v-model="search.egressType" placeholder="全部" clearable @change="fetchData" style="width: 160px">
            <el-option label="本机直连" value="direct" />
            <el-option label="出站服务" value="egress" />
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
        <el-table-column prop="proxyName" label="代理名称" min-width="180" align="center" show-overflow-tooltip />
        <el-table-column label="所属服务" min-width="160" align="center" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.serviceName ? row.serviceName + '-' + row.machineName : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="出站服务" width="140" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.egressServiceName" type="success" size="small">{{ row.egressServiceName }}</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="公网地址" min-width="180" align="center">
          <template #default="{ row }">
            {{ row.publicIp }}:{{ row.publicPort }}
          </template>
        </el-table-column>
        <el-table-column label="内网地址" min-width="180" align="center">
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
        <el-table-column label="操作" width="200" fixed="right" align="center">
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
          <el-select v-model="form.serviceId" placeholder="请选择" style="width: 100%" filterable :teleported="false" @change="onServiceChange">
            <el-option-group label="Docker服务">
              <el-option v-for="s in dockerServiceOptions" :key="'docker-' + s.id" :label="s.name + '-' + s.machineName" :value="'docker-' + s.id" />
            </el-option-group>
            <el-option-group label="其他服务">
              <el-option v-for="s in otherServiceOptions" :key="'other-' + s.id" :label="s.name + '-' + s.machineName" :value="'other-' + s.id" />
            </el-option-group>
          </el-select>
        </el-form-item>
        <el-form-item label="出站服务" prop="isDirect">
          <el-radio-group v-model="form.isDirect" @change="onEgressTypeChange">
            <el-radio :value="true">本机直连</el-radio>
            <el-radio :value="false">出站服务</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="!form.isDirect" label="选择出站服务" prop="egressServiceId">
          <el-select v-model="form.egressServiceId" placeholder="请选择" style="width: 100%" filterable :teleported="false" @change="onEgressServiceChange">
            <el-option v-for="s in egressServiceOptions" :key="s.id" :label="s.name + '-' + s.machineName" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="代理/隧道名称">
          <el-input v-model="form.proxyName" placeholder="例如：nginx-frp-tunnel" />
        </el-form-item>
        <el-form-item label="公网IP" prop="publicIp">
          <el-input v-model="form.publicIp" placeholder="对外公网IP" />
        </el-form-item>
        <el-form-item label="公网端口" prop="publicPort">
          <el-input-number v-model="form.publicPort" :min="0" :max="65535" style="width: 100%" />
        </el-form-item>
        <el-form-item label="内网IP" prop="internalIp">
          <el-input v-model="form.internalIp" placeholder="内网IP地址" />
        </el-form-item>
        <el-form-item label="内网端口" prop="internalPort">
          <el-input-number v-model="form.internalPort" :min="0" :max="65535" style="width: 100%" />
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
import { getOtherServices } from '../api/otherService'
import { ElMessage, ElMessageBox } from 'element-plus'

const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const loading = ref(false)
const dockerServiceOptions = ref([])
const otherServiceOptions = ref([])
const egressServiceOptions = ref([])
const allServiceOptions = ref([])
const search = reactive({ serviceId: '', egressType: '', status: '' })
const formVisible = ref(false)
const formMode = ref('create')
const formRef = ref(null)
const submitting = ref(false)
const editId = ref(null)

const form = reactive({
  serviceId: '', serviceType: 'docker', isDirect: true, egressServiceId: '', methodType: 'FRP', proxyName: '',
  publicIp: '', publicPort: 0, internalIp: '', internalPort: 0,
  protocol: 'TCP', status: 1, remark: ''
})

const rules = {
  serviceId: [{ required: true, message: '请选择所属服务', trigger: 'change' }],
  isDirect: [{ required: true, message: '请选择出站服务类型', trigger: 'change' }],
  egressServiceId: [{ required: true, message: '请选择出站服务', trigger: 'change' }],
  publicIp: [{ required: true, message: '请输入公网IP', trigger: 'blur' }],
  publicPort: [{ required: true, message: '请输入公网端口', trigger: 'blur' }],
  internalIp: [{ required: true, message: '请输入内网IP', trigger: 'blur' }],
  internalPort: [{ required: true, message: '请输入内网端口', trigger: 'blur' }]
}

const fetchData = async () => {
  loading.value = true
  try {
    const params = { page: page.value, pageSize: pageSize.value }
    if (search.serviceId) params.serviceId = search.serviceId
    if (search.egressType === 'direct') params.isDirect = true
    else if (search.egressType === 'egress') params.isDirect = false
    if (search.status !== '') params.status = search.status
    const res = await getEgressMethods(params)
    list.value = res.data.list
    total.value = res.data.total
  } catch {
  } finally {
    loading.value = false
  }
}

const fetchServices = async () => {
  try {
    const [dockerRes, otherRes] = await Promise.all([
      getServices({ page: 1, pageSize: 200 }),
      getOtherServices({ page: 1, pageSize: 200 })
    ])
    dockerServiceOptions.value = dockerRes.data.list
    otherServiceOptions.value = otherRes.data.list
    egressServiceOptions.value = dockerRes.data.list.filter(s => s.isEgress)

    const options = []
    for (const s of dockerRes.data.list) {
      options.push({ value: 'docker-' + s.id, label: s.name + '-' + (s.machineName || '') })
    }
    for (const s of otherRes.data.list) {
      options.push({ value: 'other-' + s.id, label: s.name + '-' + (s.machineName || '') })
    }
    allServiceOptions.value = options
  } catch {}
}

const getServiceInfo = (serviceIdStr) => {
  if (!serviceIdStr) return null
  if (serviceIdStr.startsWith('docker-')) {
    const id = parseInt(serviceIdStr.replace('docker-', ''))
    const s = dockerServiceOptions.value.find(x => x.id === id)
    return s ? { ...s, serviceType: 'docker', numericId: id } : null
  } else if (serviceIdStr.startsWith('other-')) {
    const id = parseInt(serviceIdStr.replace('other-', ''))
    const s = otherServiceOptions.value.find(x => x.id === id)
    return s ? { ...s, serviceType: 'other', numericId: id } : null
  }
  return null
}

const onServiceChange = () => {
  autoFillAddresses()
}

const onEgressTypeChange = () => {
  form.egressServiceId = ''
  autoFillAddresses()
}

const onEgressServiceChange = () => {
  autoFillAddresses()
}

const autoFillAddresses = () => {
  const svc = getServiceInfo(form.serviceId)
  if (!svc) return

  const serviceMachineIp = svc.machineIp || ''
  const servicePort = svc.port || 0

  if (form.isDirect) {
    form.publicIp = serviceMachineIp
    form.publicPort = servicePort
    form.internalIp = serviceMachineIp
    form.internalPort = servicePort
  } else {
    const egressId = Number(form.egressServiceId)
    if (egressId > 0) {
      const egressSvc = egressServiceOptions.value.find(x => x.id === egressId)
      if (egressSvc) {
        form.publicIp = egressSvc.machineIp || ''
        form.publicPort = egressSvc.port || 0
        form.internalIp = serviceMachineIp
        form.internalPort = servicePort
      }
    }
  }
}

const openForm = (mode, row) => {
  formMode.value = mode
  formVisible.value = true
  if (mode === 'edit' && row) {
    editId.value = row.id
    const sid = row.serviceType === 'other' ? 'other-' + row.serviceId : 'docker-' + row.serviceId
    Object.assign(form, {
      serviceId: sid, serviceType: row.serviceType || 'docker',
      isDirect: row.isDirect || false, egressServiceId: row.egressServiceId || '', methodType: row.methodType, proxyName: row.proxyName || '',
      publicIp: row.publicIp, publicPort: row.publicPort,
      internalIp: row.internalIp, internalPort: row.internalPort,
      protocol: row.protocol || 'TCP', status: row.status, remark: row.remark || ''
    })
  } else {
    editId.value = null
    Object.assign(form, {
      serviceId: '', serviceType: 'docker', isDirect: true, egressServiceId: '', methodType: 'FRP', proxyName: '',
      publicIp: '', publicPort: 0, internalIp: '', internalPort: 0,
      protocol: 'TCP', status: 1, remark: ''
    })
  }
}

const handleSubmit = async () => {
  if (!form.isDirect && !form.egressServiceId) {
    ElMessage.warning('请选择出站服务')
    return
  }
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    const svc = getServiceInfo(form.serviceId)
    const payload = {
      serviceId: svc ? svc.numericId : 0,
      serviceType: svc ? svc.serviceType : 'docker',
      isDirect: form.isDirect,
      egressServiceId: form.isDirect ? 0 : Number(form.egressServiceId),
      methodType: form.methodType,
      proxyName: form.proxyName,
      publicIp: form.publicIp,
      publicPort: form.publicPort,
      internalIp: form.internalIp,
      internalPort: form.internalPort,
      protocol: form.protocol,
      status: form.status,
      remark: form.remark
    }
    if (formMode.value === 'create') {
      await createEgressMethod(payload)
      ElMessage.success('创建成功')
    } else {
      await updateEgressMethod(editId.value, payload)
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
  const proto = (row.protocol || '').toUpperCase()
  let addr
  if (proto === 'HTTP' || proto === 'HTTPS') {
    addr = `${proto.toLowerCase()}://${row.publicIp}:${row.publicPort}`
  } else {
    addr = `${row.publicIp}:${row.publicPort}`
  }
  if (navigator.clipboard && window.isSecureContext) {
    navigator.clipboard.writeText(addr).then(() => {
      ElMessage.success(`已复制: ${addr}`)
    }).catch(() => {
      fallbackCopy(addr)
    })
  } else {
    fallbackCopy(addr)
  }
}

const fallbackCopy = (text) => {
  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.style.position = 'fixed'
  textarea.style.left = '-9999px'
  document.body.appendChild(textarea)
  textarea.select()
  try {
    document.execCommand('copy')
    ElMessage.success(`已复制: ${text}`)
  } catch {
    ElMessage.warning('复制失败，请手动复制')
  }
  document.body.removeChild(textarea)
}

onMounted(() => {
  fetchData()
  fetchServices()
})
</script>
