import request from './request'

export const getServices = (params) => request.get('/docker-services', { params })
export const createService = (data) => request.post('/docker-services', data)
export const updateService = (id, data) => request.put(`/docker-services/${id}`, data)
export const deleteService = (id) => request.delete(`/docker-services/${id}`)
export const checkService = (id) => request.post(`/docker-services/${id}/check`)