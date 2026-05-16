import request from './request'

export const exportConfig = () => request.get('/config/export')
export const importConfig = (data) => request.post('/config/import', data)
