import axios from 'axios';

const api = axios.create({
  baseURL: '/api',
  timeout: 30000,
});

// 新闻
export const getNews = (params?: { category?: string; source?: string; limit?: number; offset?: number }) =>
  api.get('/news', { params });
export const getNewsDetail = (id: string) => api.get(`/news/${id}`);
export const deleteNews = (id: string) => api.delete(`/news/${id}`);
export const triggerCollect = () => api.post('/news/collect');
export const triggerProcess = () => api.post('/news/process');

// 阅读窗口
export const getReadingNews = (params?: { category?: string; pushed?: string; limit?: number; offset?: number }) =>
  api.get('/reading', { params });
export const addToReading = (id: string) => api.post(`/reading/${id}/add`);
export const removeFromReading = (id: string) => api.post(`/reading/${id}/remove`);
export const clearPushedNews = () => api.post('/reading/clear-pushed');

// 新闻源
export const getSources = () => api.get('/sources');
export const createSource = (data: any) => api.post('/sources', data);
export const updateSource = (id: string, data: any) => api.put(`/sources/${id}`, data);
export const deleteSource = (id: string) => api.delete(`/sources/${id}`);

// 推送渠道
export const getChannels = () => api.get('/channels');
export const createChannel = (data: any) => api.post('/channels', data);
export const updateChannel = (id: string, data: any) => api.put(`/channels/${id}`, data);
export const deleteChannel = (id: string) => api.delete(`/channels/${id}`);
export const testChannel = (id: string) => api.post(`/channels/${id}/test`);

// 推送任务
export const getTasks = () => api.get('/tasks');
export const createTask = (data: any) => api.post('/tasks', data);
export const updateTask = (id: string, data: any) => api.put(`/tasks/${id}`, data);
export const deleteTask = (id: string) => api.delete(`/tasks/${id}`);
export const runTask = (id: string) => api.post(`/tasks/${id}/run`);

// 邮件模板
export const getTemplates = () => api.get('/templates');
export const createTemplate = (data: any) => api.post('/templates', data);
export const updateTemplate = (id: string, data: any) => api.put(`/templates/${id}`, data);
export const deleteTemplate = (id: string) => api.delete(`/templates/${id}`);
export const previewTemplate = (content: string) => api.post('/templates/preview', { content });

// AI配置
export const getAIConfig = () => api.get('/ai/config');
export const saveAIConfig = (data: any) => api.post('/ai/config', data);
export const translateText = (text: string, targetLang?: string) => api.post('/ai/translate', { text, target_lang: targetLang });
export const summarizeText = (text: string) => api.post('/ai/summarize', { text });

// 统计
export const getStats = () => api.get('/stats');

// 自动打包推送
export const getAutoPushConfig = () => api.get('/auto-push/config');
export const saveAutoPushConfig = (data: { enabled: boolean; threshold: number; channel_id: string; template_id: string }) => 
  api.post('/auto-push/config', data);
export const getAutoPushStatus = () => api.get('/auto-push/status');

export default api;
