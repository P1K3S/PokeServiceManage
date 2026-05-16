import request from './request'

export const getNotice = () => request.get('/notices')
export const updateNotice = (data) => request.put('/notices', data)
