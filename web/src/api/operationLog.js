import request from './request'

export const getOperationLogs = (params) => request.get('/operation-logs', { params })
