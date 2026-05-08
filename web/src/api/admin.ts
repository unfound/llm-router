const BASE_URL = '';

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${url}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || '请求失败');
  }
  return res.json();
}

// 统计
export const getStatsOverview = () => request<any>('/admin/api/stats/overview');
export const getStatsModels = () => request<any>('/admin/api/stats/models');
export const getStatsTimeSeries = (hours = 24) => request<any>(`/admin/api/stats/timeseries?hours=${hours}`);

// 端点
export const getEndpoints = () => request<any>('/admin/api/endpoints');
export const createEndpoint = (data: any) => request<any>('/admin/api/endpoints', { method: 'POST', body: JSON.stringify(data) });
export const deleteEndpoint = (id: number) => request<any>(`/admin/api/endpoints/${id}`, { method: 'DELETE' });

// 模型
export const getModels = () => request<any>('/admin/api/models');
export const createModel = (data: any) => request<any>('/admin/api/models', { method: 'POST', body: JSON.stringify(data) });
export const updateModel = (id: number, data: any) => request<any>(`/admin/api/models/${id}`, { method: 'PUT', body: JSON.stringify(data) });
export const deleteModel = (id: number) => request<any>(`/admin/api/models/${id}`, { method: 'DELETE' });
export const toggleModel = (id: number) => request<any>(`/admin/api/models/${id}/toggle`, { method: 'PUT' });
export const syncModels = () => request<any>('/admin/api/models/sync', { method: 'POST' });

// 日志
export const getLogs = (params?: { session_id?: string; model_name?: string; status?: string; limit?: number; offset?: number }) => {
  const query = new URLSearchParams();
  if (params?.session_id) query.set('session_id', params.session_id);
  if (params?.model_name) query.set('model_name', params.model_name);
  if (params?.status) query.set('status', params.status);
  if (params?.limit) query.set('limit', String(params.limit));
  if (params?.offset) query.set('offset', String(params.offset));
  return request<any>(`/admin/api/logs?${query.toString()}`);
};
export const getLogDetail = (id: number) => request<any>(`/admin/api/logs/${id}`);

// 会话
export const getSessions = () => request<any>('/admin/api/sessions');
