import request from './request'

export const getEgressMethods = (params) => request.get('/egress-methods', { params })
export const createEgressMethod = (data) => request.post('/egress-methods', data)
export const updateEgressMethod = (id, data) => request.put(`/egress-methods/${id}`, data)
export const deleteEgressMethod = (id) => request.delete(`/egress-methods/${id}`)