<template>
  <div>
    <el-card shadow="hover">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>主机列表</span>
          <el-button type="primary" @click="openForm('create')">新增主机</el-button>
        </div>
      </template>

      <el-form :inline="true" :model="search" style="margin-bottom: 10px">
        <el-form-item label="名称">
          <el-input v-model="search.keyword" placeholder="模糊搜索" clearable @clear="fetchData" @keyup.enter="fetchData" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="search.machineType" placeholder="全部" clearable @change="fetchData" style="width: 120px">
            <el-option label="局域网" value="LAN" />
            <el-option label="云服务器" value="CLOUD" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="search.status" placeholder="全部" clearable @change="fetchData" style="width: 120px">
            <el-option label="在线" :value="1" />
            <el-option label="离线" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchData">查询</el-button>
          <el-button @click="handleCheckAllSSH" :loading="allSshChecking" style="margin-left: 8px">连通检测</el-button>
          <el-button @click="handleDiscoverAllServices" :loading="allDiscovering" style="margin-left: 8px">Docker服务检测</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="list" stripe border v-loading="loading" @expand-change="onExpandChange">
        <el-table-column type="expand" width="40">
          <template #default="{ row }">
            <el-table :data="row._services" stripe border size="small" v-if="row._services && row._services.length > 0">
              <el-table-column prop="name" label="服务名称" />
              <el-table-column prop="type" label="类型" width="100" />
              <el-table-column prop="port" label="端口" width="80" align="center" />
              <el-table-column prop="protocol" label="协议" width="70" align="center" />
              <el-table-column label="状态" width="70" align="center">
                <template #default="{ row: sr }">
                  <el-tag :type="sr.status === 1 ? 'primary' : 'info'" size="small">
                    {{ sr.status === 1 ? '运行中' : '已停止' }}
                  </el-tag>
                </template>
              </el-table-column>
            </el-table>
            <el-empty description="暂无服务" v-else />
          </template>
        </el-table-column>
        <el-table-column prop="name" label="主机名称" min-width="140" />
        <el-table-column prop="ip" label="IP地址" width="150" />
        <el-table-column label="类型" width="90">
          <template #default="{ row }">
            <el-tag :type="row.machineType === 'CLOUD' ? '' : 'success'" size="small">
              {{ row.machineType === 'CLOUD' ? '云服务器' : '局域网' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="cpu" label="CPU" min-width="120" />
        <el-table-column prop="memory" label="内存" width="80" />
        <el-table-column prop="os" label="操作系统" min-width="120" />
        <el-table-column label="状态" width="70" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '在线' : '离线' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="serviceCount" label="服务数" width="70" align="center" />
        <el-table-column label="操作" width="170">
          <template #default="{ row }">
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

    <el-dialog v-model="formVisible" :title="formMode === 'create' ? '新增主机' : '编辑主机'" width="560px" :close-on-click-modal="false" :lock-scroll="false">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="主机名称" prop="name">
          <el-input v-model="form.name" placeholder="例如：办公内网-Node1" />
        </el-form-item>
        <el-form-item label="IP地址" prop="ip">
          <el-input v-model="form.ip" placeholder="管理IP地址" />
        </el-form-item>
        <el-form-item label="类型" prop="machineType">
          <el-select v-model="form.machineType" placeholder="请选择" style="width: 100%">
            <el-option label="局域网" value="LAN" />
            <el-option label="云服务器" value="CLOUD" />
          </el-select>
        </el-form-item>
        <el-form-item label="CPU">
          <el-input v-model="form.cpu" placeholder="例如：Intel i7-12700" />
        </el-form-item>
        <el-form-item label="内存">
          <el-input v-model="form.memory" placeholder="例如：32GB" />
        </el-form-item>
        <el-form-item label="磁盘">
          <el-input v-model="form.disk" placeholder="例如：1TB SSD" />
        </el-form-item>
        <el-form-item label="操作系统">
          <el-input v-model="form.os" placeholder="例如：Ubuntu 22.04" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" active-text="在线" inactive-text="离线" />
        </el-form-item>
        <el-divider content-position="left">SSH 连接信息</el-divider>
        <el-form-item label="SSH端口">
          <el-input-number v-model="form.sshPort" :min="1" :max="65535" style="width: 100%" placeholder="默认 22" />
        </el-form-item>
        <el-form-item label="SSH用户">
          <el-input v-model="form.sshUser" placeholder="默认 root" />
        </el-form-item>
        <el-form-item label="SSH密码">
          <el-input v-model="form.sshPassword" type="password" show-password placeholder="登录密码" />
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
import { getMachines, createMachine, updateMachine, deleteMachine, checkMachineSSH, discoverMachineServices } from '../api/machine'
import { getServices } from '../api/service'
import { ElMessage, ElMessageBox } from 'element-plus'

const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const loading = ref(false)
const search = reactive({ keyword: '', machineType: '', status: '' })
const formVisible = ref(false)
const formMode = ref('create')
const formRef = ref(null)
const submitting = ref(false)
const allSshChecking = ref(false)
const allDiscovering = ref(false)
const editId = ref(null)

const form = reactive({
  name: '', ip: '', machineType: 'LAN', cpu: '', memory: '', disk: '', os: '',
  sshPort: 22, sshUser: 'root', sshPassword: '', status: 1, remark: ''
})

const rules = {
  name: [{ required: true, message: '请输入主机名称', trigger: 'blur' }],
  ip: [{ required: true, message: '请输入IP地址', trigger: 'blur' }],
  machineType: [{ required: true, message: '请选择主机类型', trigger: 'change' }]
}

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getMachines({ page: page.value, pageSize: pageSize.value, ...search })
    list.value = res.data.list
    total.value = res.data.total
  } catch {
  } finally {
    loading.value = false
  }
}

const onExpandChange = async (row) => {
  if (row._services) return
  try {
    const res = await getServices({ machineId: row.id, page: 1, pageSize: 100 })
    row._services = res.data.list
  } catch {
    row._services = []
  }
}

const openForm = (mode, row) => {
  formMode.value = mode
  formVisible.value = true
  if (mode === 'edit' && row) {
    editId.value = row.id
    Object.assign(form, {
      name: row.name, ip: row.ip, machineType: row.machineType,
      cpu: row.cpu || '', memory: row.memory || '', disk: row.disk || '',
      os: row.os || '', sshPort: row.sshPort || 22, sshUser: row.sshUser || 'root',
      sshPassword: '', status: row.status, remark: row.remark || ''
    })
  } else {
    editId.value = null
    Object.assign(form, { name: '', ip: '', machineType: 'LAN', cpu: '', memory: '', disk: '', os: '', sshPort: 22, sshUser: 'root', sshPassword: '', status: 1, remark: '' })
  }
}

const handleCheckSSH = async (row) => {
  row._sshChecking = true
  try {
    const res = await checkMachineSSH(row.id)
    row.status = res.data.status
    ElMessage.success(res.data.message)
    fetchData()
  } catch {
    ElMessage.error('检测失败')
  } finally {
    row._sshChecking = false
  }
}

const handleDiscoverServices = async (row) => {
  row._discovering = true
  try {
    const res = await discoverMachineServices(row.id)
    ElMessage.success(res.data.message)
    fetchData()
  } catch (e) {
    const msg = e.response?.data?.message || '检测失败'
    ElMessage.error(msg)
  } finally {
    row._discovering = false
  }
}

const handleCheckAllSSH = async () => {
  allSshChecking.value = true
  let online = 0
  let offline = 0
  for (const m of list.value) {
    try {
      const res = await checkMachineSSH(m.id)
      m.status = res.data.status
      if (res.data.status === 1) {
        online++
      } else {
        offline++
      }
    } catch {
      offline++
    }
  }
  ElMessage.success(`检测完成：在线 ${online} 个，离线 ${offline} 个`)
  fetchData()
  allSshChecking.value = false
}

const handleDiscoverAllServices = async () => {
  allDiscovering.value = true
  let total = 0
  for (const m of list.value) {
    try {
      const res = await discoverMachineServices(m.id)
      total += res.data.count || 0
    } catch {
    }
  }
  ElMessage.success(`检测完成：更新 ${total} 个 Docker 服务`)
  fetchData()
  allDiscovering.value = false
}

const handleSubmit = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    if (formMode.value === 'create') {
      await createMachine(form)
      ElMessage.success('创建成功')
    } else {
      await updateMachine(editId.value, form)
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
  ElMessageBox.confirm(`确定删除主机「${row.name}」吗？该主机下的所有服务和出站方式也将被删除。`, '确认删除', {
    confirmButtonText: '确定删除',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    await deleteMachine(row.id)
    ElMessage.success('删除成功')
    fetchData()
  }).catch(() => {})
}

onMounted(fetchData)
</script>