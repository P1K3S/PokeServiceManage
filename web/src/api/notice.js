import request from './request'

export const getNotices = () => request.get('/notices')
export const createNotice = (data) => request.post('/notices', data)
export const updateNotice = (id, data) => request.put(`/notices/${id}`, data)
export const deleteNotice = (id) => request.delete(`/notices/${id}`)
export const togglePinNotice = (id) => request.put(`/notices/${id}/pin`)
export const moveNotice = (id, direction) => request.put(`/notices/${id}/move/${direction}`)
