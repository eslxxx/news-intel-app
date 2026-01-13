import React, { useState, useEffect } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Layout as AntLayout, Menu, Button, Drawer } from 'antd';
import {
  DashboardOutlined,
  ReadOutlined,
  GlobalOutlined,
  SendOutlined,
  ScheduleOutlined,
  FileTextOutlined,
  RobotOutlined,
  MenuOutlined,
  BookOutlined,
} from '@ant-design/icons';

const { Sider, Content, Header } = AntLayout;

const Layout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [isMobile, setIsMobile] = useState(window.innerWidth < 768);
  const [drawerVisible, setDrawerVisible] = useState(false);

  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768);
    };
    
    window.addEventListener('resize', checkMobile);
    return () => window.removeEventListener('resize', checkMobile);
  }, []);

  const menuItems = [
    { key: '/dashboard', icon: <DashboardOutlined />, label: '仪表盘' },
    { key: '/reading', icon: <BookOutlined />, label: '阅读窗口' },
    { key: '/news', icon: <ReadOutlined />, label: '全部新闻' },
    { key: '/sources', icon: <GlobalOutlined />, label: '新闻源' },
    { key: '/channels', icon: <SendOutlined />, label: '推送渠道' },
    { key: '/tasks', icon: <ScheduleOutlined />, label: '推送任务' },
    { key: '/templates', icon: <FileTextOutlined />, label: '邮件模板' },
    { key: '/ai', icon: <RobotOutlined />, label: 'AI 设置' },
  ];

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key);
    setDrawerVisible(false);
  };

  const menuContent = (
    <Menu
      theme="dark"
      mode="inline"
      selectedKeys={[location.pathname]}
      items={menuItems}
      onClick={handleMenuClick}
      style={{ borderRight: 0 }}
    />
  );

  if (isMobile) {
    return (
      <AntLayout className="mobile-layout">
        <Header className="mobile-header">
          <Button
            type="text"
            icon={<MenuOutlined style={{ fontSize: 20, color: '#fff' }} />}
            onClick={() => setDrawerVisible(true)}
            className="menu-trigger"
          />
          <span className="mobile-title">News Intel</span>
        </Header>
        <Drawer
          placement="left"
          onClose={() => setDrawerVisible(false)}
          open={drawerVisible}
          width={240}
          styles={{ 
            body: { padding: 0, background: '#001529' },
            header: { display: 'none' }
          }}
        >
          <div className="drawer-logo">News Intel</div>
          {menuContent}
        </Drawer>
        <Content className="mobile-content">
          <Outlet />
        </Content>
      </AntLayout>
    );
  }

  return (
    <AntLayout style={{ minHeight: '100vh' }}>
      <Sider theme="dark" width={220} style={{ position: 'fixed', left: 0, top: 0, bottom: 0 }}>
        <div className="logo">News Intel</div>
        {menuContent}
      </Sider>
      <AntLayout style={{ marginLeft: 220 }}>
        <Content style={{ margin: 24, padding: 24, background: '#f5f5f5', borderRadius: 8, minHeight: 'calc(100vh - 48px)' }}>
          <Outlet />
        </Content>
      </AntLayout>
    </AntLayout>
  );
};

export default Layout;
