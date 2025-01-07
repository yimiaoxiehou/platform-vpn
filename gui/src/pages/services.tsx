import React, { useEffect, useState } from 'react';
import { Card, Spin, message, Tag, Layout, Input, Space, Row, Col } from 'antd';
import { GetServices } from "../../wailsjs/go/main/App";
import { main } from "../../wailsjs/go/models";

import { BrowserOpenURL, ClipboardSetText } from '../../wailsjs/runtime';
import { SearchOutlined } from '@ant-design/icons';
import { Content, Header } from 'antd/es/layout/layout';

const Services: React.FC = () => {
  const [nsServices, setNsServices] = useState<Array<main.AppNsService>>([]);
  const [initNsServices, setInitNsServices] = useState<Array<main.AppNsService>>([]);
  const [loading, setLoading] = useState<boolean>(true);

  const [isVPNActive] = useState<boolean>(() => {
    const stored = localStorage.getItem('isVPNActive');
    return stored === 'true';
  });

  useEffect(() => {
    if (!isVPNActive) {
      message.error('VPN未启动');
      return;
    }
    const fetchServices = () => {
      setLoading(true);

      GetServices().then((services) => {
        services.sort((a, b) => a.Namespace.localeCompare(b.Namespace));
        setInitNsServices(services);
        setNsServices(services);
      }).catch(() => {
          message.error('获取服务失败');
      });
      setLoading(false);
    };

    fetchServices();
  }, [isVPNActive]);

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    );
  }

  function AddToClipboard(text: string) {
    ClipboardSetText(text).then(() => {
      message.success("复制成功:" + text);
    });
  }

  function ServiceTitleEle(ns: string, service: main.AppService) {
    return (
      <Space direction="vertical" size={0}>
        <h3
          style={{ margin: 0, fontSize: '16px', fontWeight: 'bold' }}
          onClick={() => AddToClipboard(service.Name + ":" + ns)}
        >
          {service.Name}
        </h3>
        <span
          style={{ fontSize: '12px', color: 'rgba(0, 0, 0, 0.45)' }}
          onClick={() => AddToClipboard(ns)}
        >
          {ns}
        </span>
        <code
          style={{ fontSize: '12px', color: 'rgba(0, 0, 0, 0.45)' }}
          onClick={() => AddToClipboard(service.IP)}
        >
          {service.IP}
        </code>
      </Space>
    );
  }

  function filter(value: string) {
    setNsServices(structuredClone(initNsServices));
    const _nsServices: Array<main.AppNsService> = [];
    initNsServices.forEach((nsServices) => {
      if (nsServices.Namespace.includes(value)) {
        _nsServices.push(nsServices);
        return;
      }
      let _service = Array.from(nsServices.Services);
      _service = _service.filter((service) => service.Name.includes(value));
      if (_service.length > 0) {
        nsServices.Services = _service;
        _nsServices.push(nsServices);
      }
    });
    setNsServices(_nsServices);
  }

  const serviceCardStyle = {
    body: {
      padding: '12px',
    },
    header: {
      padding: '12px',
      minHeight: '80px',
    },
  };

  return (
    <Layout style={{ height: '100vh', overflow: 'hidden' }}>
      <Header
        style={{
          padding: '0 24px',
          background: '#fff',
          display: 'flex',
          alignItems: 'center',
          height: '56px',
          boxShadow: '0 1px 2px rgba(0, 0, 0, 0.03)',
        }}
      >
        <Input
          placeholder="搜索服务..."
          allowClear
          prefix={<SearchOutlined style={{ color: '#00000040' }} />}
          style={{
            maxWidth: 300,
            borderRadius: 4,
          }}
          onChange={(e) => filter(e.target.value)}
        />
      </Header>
      <Content
        style={{
          padding: '16px',
          overflow: 'auto',
          height: 'calc(100vh - 56px)',
        }}
      >
        <Row gutter={[16, 16]}>
          {nsServices.map((nsService) =>
            nsService.Services.filter((service) => service.IP !== "None").map((service) => (
              <Col xs={24} sm={12} md={12} lg={8} xl={6} key={service.Name}>
                <Card
                  type="inner"
                  size="small"
                  styles={serviceCardStyle}
                  title={ServiceTitleEle(nsService.Namespace, service)}
                >
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: '4px' }}>
                    {service.Ports.map((port, index) => (
                      <Tag
                        key={`${service.Name}-port-${index}`}
                        style={{
                          backgroundColor: '#f0f5ff',
                          border: '1px solid #d6e4ff',
                          color: '#4096ff',
                          fontFamily: 'monospace',
                          fontSize: '12px',
                          margin: 0,
                          cursor: 'pointer',
                        }}
                        onClick={() => {
                          AddToClipboard(`${service.IP}:${port}`);
                          BrowserOpenURL(`http://${service.IP}:${port}`);
                        }}
                      >
                        :{port}
                      </Tag>
                    ))}
                  </div>
                </Card>
              </Col>
            ))
          )}
        </Row>
      </Content>
    </Layout>
  );
};

export default Services;