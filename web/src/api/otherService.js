import request from './request'

export const getOtherServices = (params) => request.get('/other-services', { params })
export const createOtherService = (data) => request.post('/other-services', data)
export const updateOtherService = (id, data) => request.put(`/other-services/${id}`, data)
export const deleteOtherService = (id) => request.delete(`/other-services/${id}`)