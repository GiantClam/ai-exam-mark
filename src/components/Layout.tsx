import React from 'react';
import { Layout as AntLayout, Menu } from 'antd';
import { HomeOutlined, UploadOutlined, CheckSquareOutlined } from '@ant-design/icons';
import Link from 'next/link';

const { Header, Content } = AntLayout;

const menuItems = [
  {
    key: 'home',
    icon: <HomeOutlined />,
    label: <Link href="/">首页</Link>,
  },
  {
    key: 'homework',
    icon: <UploadOutlined />,
    label: <Link href="/homework">作业上传</Link>,
  },
  {
    key: 'mark',
    icon: <CheckSquareOutlined />,
    label: <Link href="/MarkHomework">作业批改</Link>,
  },
];

const Layout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return (
    <AntLayout style={{ minHeight: '100vh' }}>
      <Header>
        <Menu
          theme="dark"
          mode="horizontal"
          defaultSelectedKeys={['home']}
          items={menuItems}
          style={{ lineHeight: '64px' }}
        />
      </Header>
      <Content style={{ padding: '24px', background: '#fff' }}>
        {children}
      </Content>
    </AntLayout>
  );
};

export default Layout; 