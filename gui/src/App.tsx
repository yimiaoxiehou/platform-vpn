import React from 'react';
import './index.css';
import { ConfigProvider, Layout, Menu } from 'antd';
import { Route, Routes, useNavigate } from 'react-router-dom';
import Log from './pages/log';
import { Home } from './pages/home';
import Services from './pages/services';
import { AppstoreOutlined, FileTextOutlined, HomeOutlined } from '@ant-design/icons';
const { Sider, Content } = Layout;


const App: React.FC = () => {
  const navigate = useNavigate();

  return (
    <ConfigProvider
        theme={{
          cssVar: true,
          components: {
            Menu: {
              darkItemSelectedBg: '#57606f',
            },
            Input: {
              hoverBorderColor: '#57606f',
              activeBorderColor: '#57606f',
            },
            Select: {
              hoverBorderColor: '#57606f',
              activeBorderColor: '#57606f',
            },
            Button: {
              defaultActiveBorderColor: '#57606f',
              defaultHoverBorderColor: '#57606f',
            },
          },
        }}
      >
    
    <Layout style={{ width: '100vw', height: '100vh' }}>
      <Sider
        width={200}
        style={{
          background: 'linear-gradient(135deg,#bbd2c5,#536976,#292e49)',
          height: '100vh',
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
          padding: '16px',
          textAlign: 'right',
        }}
      >
        <div style={{display: 'flex', flexDirection: 'column', justifyContent: 'space-between'}}>
          <h2 className="mb-2 px-4 text-lg font-semibold" style={{ color: 'white' }}>Platform VPN</h2>
          <Menu
            theme="dark"
            style={{ background: 'transparent', fontWeight: 'bold' }}
            defaultSelectedKeys={['home']}
            items={[
              {
                key: 'home',
                label: (
                  <span>
                    首页&nbsp;&nbsp;<HomeOutlined />
                  </span>
                ),
              },
              {
                key: 'services',
                label: (
                  <span>
                    服务列表&nbsp;&nbsp;<AppstoreOutlined />
                  </span>
                ),
              },
              {
                key: 'log',
                label: (
                  <span>
                    日志&nbsp;&nbsp;<FileTextOutlined />
                  </span>
                ),
              }
            ]}
            onSelect={({ key }) => {
              navigate(key);
            }}
          />
        </div>
       <span style={{color: '#dfe4ea', fontWeight: 'bold'}}>version: 1.0.0<br/>2025-01-17 10:20:22</span>
      </Sider>
      <Layout style={{ marginLeft: 200 }} >
        <Content
          className='rollable'
          style={{
            minHeight: 'calc(100vh - 48px)',
            overflow: 'auto',
          }}
        >
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/home" element={<Home />} />
            <Route path="/services" element={<Services />} />
            <Route path="/log" element={<Log />} />
          </Routes>
          </Content>
        </Layout>
      </Layout>
    </ConfigProvider>
  );
};

export default App;