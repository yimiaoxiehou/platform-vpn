import { useState, useEffect } from 'react';
import { Button, Form, Input, Select, Row, Col, message } from 'antd';
import { OpenHosts, RefreshHosts, StartVPN, StopVPN } from "../../wailsjs/go/main/App";
import { main } from "../../wailsjs/go/models";
export const Home = () => {

  const [config, setConfig] = useState<main.VPNConfig>(() => {
    const storedConfig = localStorage.getItem('vpnConfig');
    return storedConfig
      ? new main.VPNConfig(JSON.parse(storedConfig))
      : new main.VPNConfig({
          Server: '',
          User: 'root',
          Password: '',
          RefreshInterval: 1,
        });
  });

  useEffect(() => {
    localStorage.setItem('vpnConfig', JSON.stringify(config));
  }, [config]);



  const [isVPNActive, setIsVPNActive] = useState<boolean>(() => {
    return localStorage.getItem('isVPNActive') === 'true';
  });

  useEffect(() => {
    localStorage.setItem('isVPNActive', String(isVPNActive));
  }, [isVPNActive]);

  const [loadingToggleVPN, setLoadingToggleVPN] = useState<boolean>(false);

  const refreshIntervalOptions = [
    { label: '1分钟', value: 1 },
    { label: '5分钟', value: 5 },
    { label: '10分钟', value: 10 },
    { label: '15分钟', value: 15 },
    { label: '30分钟', value: 30 },
  ];

  const handleChange = (field: keyof main.VPNConfig) => (value: string) => {
    setConfig((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  const handleDropdownChange = (value: number) => {
    setConfig((prev) => ({
      ...prev,
      RefreshInterval: value,
    }));
  };

  const toggleOpenHosts = () => {
    OpenHosts()
      .then(() => {
        message.success('成功打开hosts文件');
      })
      .catch(err => {
        message.error('打开hosts文件失败: ' + err);
      });
  };

  const toggleVPN = () => {
    if (loadingToggleVPN) return;
    setLoadingToggleVPN(true);
    if (isVPNActive) {
      StopVPN()
        .then(() => {
          message.success('VPN 已停止');
          setIsVPNActive(false);
        })
        .catch(err => {
          message.error('停止 VPN 失败: ' + err);
        })
        .finally(() => {
          setLoadingToggleVPN(false);
        });
    } else {
      if (!config.Server || !config.User || !config.Password || !config.RefreshInterval) {
        message.error('请填写完整配置');
        setLoadingToggleVPN(false);
        return;
      }
      StartVPN(config)
        .then((rs) => {
          console.log(rs);
          if (rs) {
            message.success('VPN 已启动');
            setIsVPNActive(true);
          } else {
            message.error('VPN 启动失败');
          }
        })
        .catch(err => {
          message.error('启动 VPN 失败: ' + err);
        })
        .finally(() => {
          console.log('finally');
          setLoadingToggleVPN(false);
        });
    }
  };

  const [loadingRefreshHosts, setLoadingRefreshHosts] = useState<boolean>(false);

  const toggleRefreshHosts = () => {
    if (loadingRefreshHosts) return;
    setLoadingRefreshHosts(true);
    RefreshHosts()
      .then(() => {
        message.success('hosts 已刷新');
      })
      .catch(err => {
        message.error('刷新 hosts 失败: ' + err);
      })
      .finally(() => {
        setLoadingRefreshHosts(false);
      });
  };

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
      <Form layout="vertical" style={{ maxWidth: '600px', padding: '15px', gap: '15px' }}>
        <div style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
          <Form.Item label="服务器">
            <Input
              allowClear
              disabled={isVPNActive}
              style={{ width: '200px' }}
              value={config.Server}
              onChange={(e) => handleChange('Server')(e.target.value)}
            />
        </Form.Item>
        <Form.Item label="账户">
          <Input
            allowClear
            disabled={isVPNActive}
            style={{ width: '200px' }}
            value={config.User}
            onChange={(e) => handleChange('User')(e.target.value)}
          />
        </Form.Item>
        <Form.Item label="密码">
          <Input.Password
            allowClear
            disabled={isVPNActive}
            style={{ width: '200px' }}
            value={config.Password}
            onChange={(e) => handleChange('Password')(e.target.value)}
          />
        </Form.Item>
        <Form.Item label="hosts 刷新间隔">
          <Select
            style={{ width: '200px' }}
            disabled={isVPNActive}
            value={config.RefreshInterval}
            onChange={handleDropdownChange}
            options={refreshIntervalOptions}
          />
        </Form.Item>
        </div>
        <Form.Item>
          <Row justify="center" gutter={10}>
            <Col>
              <Button
                type={isVPNActive ? "primary" : "default"}
                danger={isVPNActive}
                onClick={toggleVPN}
                loading={loadingToggleVPN}
              >
                {isVPNActive ? "停止" : "开始"}
              </Button>
            </Col>
            {isVPNActive && (
              <>
                <Col>
                  <Button onClick={toggleOpenHosts}>
                    打开hosts文件
                  </Button>
                </Col>
                <Col>
                  <Button onClick={toggleRefreshHosts} loading={loadingRefreshHosts}>
                    手动刷新 hosts
                  </Button>
                </Col>
              </>
            )}
          </Row>
        </Form.Item>
      </Form>
    </div>
  );
};
