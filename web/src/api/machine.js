import request from './request'

export const getMachines = (params) => request.get('/machines', { params })
export const createMachine = (data) => request.post('/machines', data)
export const updateMachine = (id, data) => request.put(`/machines/${id}`, data)
export const deleteMachine = (id) => request.delete(`/machines/${id}`)
export const getMachineDetail = (id) => request.get(`/machines/${id}`)
export const checkMachineSSH = (id) => request.post(`/machines/${id}/check-ssh`)
export const discoverMachineServices = (id) => request.post(`/machines/${id}/discover-services`)
export const getOverview = () => request.get('/overview')