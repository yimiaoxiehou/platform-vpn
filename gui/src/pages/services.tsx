import React, { useEffect, useState } from 'react';
import { Card, Spin, message, Tag, Layout, Input, Space, Row, Col, Button, Select, SelectProps } from 'antd';
import { GetNamespaces, GetServices } from "../../wailsjs/go/main/App";
import { main } from "../../wailsjs/go/models";

import { BrowserOpenURL, ClipboardSetText } from '../../wailsjs/runtime';
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons';
import { Content, Header } from 'antd/es/layout/layout';

const Services: React.FC = () => {
  const [nsServices, setNsServices] = useState<Array<main.AppNsService>>([]);
  const [initNsServices, setInitNsServices] = useState<Array<main.AppNsService>>([]);
  const [loading, setLoading] = useState<boolean>(true);
  // 添加 namespaces 状态
  const [namespaces, setNamespaces] = useState<SelectProps['options']>();
  // 将 filterNamespaces 也转换为状态
  const [filterNamespaces, setFilterNamespaces] = useState<string[]>(['default']);

  const [isVPNActive] = useState<boolean>(() => {
    return localStorage.getItem('isVPNActive') === 'true';
  });

  // 添加初始化方法
  const initializePage = async () => {
    if (!isVPNActive) {
      message.error('VPN未启动');
      return;
    }
    setLoading(true);
    try {
      // 获取命名空间列表
      const nsList = await GetNamespaces();
      const filteredNamespaces = nsList.filter((ns) => ns.indexOf("kube") < 0).map((ns) => {
        return {
          label: ns,
          value: ns,
        };
      });
      setNamespaces(filteredNamespaces);

      // 设置默认过滤的命名空间
      const defaultFilerNs = ["default"];
      if (nsList.includes("platform")) {
        defaultFilerNs.push("platform");
      }
      setFilterNamespaces(defaultFilerNs);
      // 获取服务列表
      const services = await GetServices();
      const filteredServices = services
      .filter((service) => defaultFilerNs.includes(service.Namespace))
      .sort((a, b) => a.Namespace.localeCompare(b.Namespace));
      setInitNsServices(filteredServices);
    } catch (error) {
      message.error('初始化失败' + error);
    } finally {
      setLoading(false);
    }
  };

  // 初始化页面
  useEffect(() => {
    initializePage();
  }, [isVPNActive]);

  const filterServices = async () => {
    // 获取服务列表
    const services = await GetServices();
    const filteredServices = services
      .filter((service) => filterNamespaces.includes(service.Namespace))
      .sort((a, b) => a.Namespace.localeCompare(b.Namespace));
    setNsServices(filteredServices);
  };

  // 命名空间变化时重新获取服务
  useEffect(() => {
    filterServices();
  }, [filterNamespaces]);

  // 修改刷新按钮的点击事件处理
  const handleRefresh = async () => {
    setLoading(true);
    try {
      const services = await GetServices();
      const filteredServices = services
        .filter((service) => filterNamespaces.includes(service.Namespace))
        .sort((a, b) => a.Namespace.localeCompare(b.Namespace));

      setInitNsServices(filteredServices);
      setNsServices(filteredServices);
      message.success('刷新成功');
    } catch (error) {
      message.error('获取服务失败' + error);
    } finally {
      setLoading(false);
    }
  };

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

  function ChangeFilterNS(nsList: string[]) {
    setFilterNamespaces(nsList);
    handleRefresh();
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
          padding: '0 24px 8px 24px',
          background: '#fff',
          display: 'flex',
          flexDirection: 'column',
          boxShadow: '0 1px 2px rgba(0, 0, 0, 0.03)',
          height: 'fit-content'
        }}
      >
        <div>
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
          <Button
            icon={<ReloadOutlined />}
            style={{
              border: 'none'
            }}
            onClick={handleRefresh}
          />
        </div>
        <Row>
          <Col flex="80px"><span>项目空间:</span></Col>
          <Col flex="auto">
            <Select
              mode="multiple"
              defaultValue={filterNamespaces}
              onChange={values => ChangeFilterNS(values)}
              options={namespaces}
              style={{ width: '100%' }}
              maxTagCount="responsive"
            />
          </Col>
        </Row>
      </Header>
      <Content
        style={{
          padding: '16px',
          overflow: 'auto',
          height: 'calc(100vh - var(--header-height, 0px))',
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