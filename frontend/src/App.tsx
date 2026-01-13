import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, App as AntdApp } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import NewsPage from './pages/NewsPage';
import ReadingPage from './pages/ReadingPage';
import SourcesPage from './pages/SourcesPage';
import ChannelsPage from './pages/ChannelsPage';
import TasksPage from './pages/TasksPage';
import TemplatesPage from './pages/TemplatesPage';
import AIConfigPage from './pages/AIConfigPage';
import './App.css';

const App: React.FC = () => {
  return (
    <ConfigProvider locale={zhCN} theme={{
      token: {
        colorPrimary: '#667eea',
        borderRadius: 8,
      },
    }}>
      <AntdApp>
        <BrowserRouter>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route index element={<Navigate to="/dashboard" replace />} />
              <Route path="dashboard" element={<Dashboard />} />
              <Route path="news" element={<NewsPage />} />
              <Route path="reading" element={<ReadingPage />} />
              <Route path="sources" element={<SourcesPage />} />
              <Route path="channels" element={<ChannelsPage />} />
              <Route path="tasks" element={<TasksPage />} />
              <Route path="templates" element={<TemplatesPage />} />
              <Route path="ai" element={<AIConfigPage />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </AntdApp>
    </ConfigProvider>
  );
};

export default App;
