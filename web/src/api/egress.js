import request from './request'

export const getEgressMethods = (params) => request.get('/egress-methods', { params })
export const createEgressMethod = (data) => request.post('/egress-methods', data)
export const updateEgressMethod = (id, data) => request.put(`/egress-methods/${id}`, data)
export const deleteEgressMethod = (id) => request.delete(`/egress-methods/${id}`)
export const syncFirewall = () => request.post('/egress-methods/sync-firewall')
export const generateFrpc = (ids) => request.post('/egress-methods/generate-frpc', { ids })