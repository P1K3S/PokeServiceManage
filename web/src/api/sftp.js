import request from './request'

export const sftpList = (machineId, path) => request.get(`/sftp/${machineId}/list`, { params: { path } })

export const sftpDownload = (machineId, path) => `/api/sftp/${machineId}/download?path=${encodeURIComponent(path)}&token=${encodeURIComponent(localStorage.getItem('token'))}`

export const sftpDownloadDir = (machineId, path) => `/api/sftp/${machineId}/download-dir?path=${encodeURIComponent(path)}&token=${encodeURIComponent(localStorage.getItem('token'))}`

export const sftpUpload = (machineId, path, file) => {
  const formData = new FormData()
  formData.append('file', file)
  return request.post(`/sftp/${machineId}/upload?path=${encodeURIComponent(path)}`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export const sftpMkdir = (machineId, path) => request.post(`/sftp/${machineId}/mkdir`, { path })

export const sftpRemove = (machineId, path, isDir) => request.delete(`/sftp/${machineId}/remove`, { data: { path, isDir } })

export const sftpRename = (machineId, oldPath, newPath) => request.put(`/sftp/${machineId}/rename`, { oldPath, newPath })

export const sftpReadFile = (machineId, path) => request.get(`/sftp/${machineId}/read`, { params: { path } })

export const sftpWriteFile = (machineId, path, content) => request.post(`/sftp/${machineId}/write`, { path, content })

export const sftpStat = (machineId, path) => request.get(`/sftp/${machineId}/stat`, { params: { path } })
