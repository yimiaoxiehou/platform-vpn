import { useEffect, useState } from 'react';
import { GetLogs } from '../../wailsjs/go/main/App';
import './log.css';
const Log = () => {
  const [logs, setLogs] = useState<string>('');

  useEffect(() => {
    const fetchLogs = async () => {
      try {
        const logItems = await GetLogs();
        const formattedLogs = logItems.map(item => {
          const date = new Date(item.Time);
          const formattedTime = date.toLocaleString('zh-CN', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
          });
          return `${formattedTime} ${item.Level} ${item.Message}`;
        }).join('\n');
        setLogs(formattedLogs);
        console.log(formattedLogs);
      } catch (error) {
        console.error('获取日志失败:', error);
      }
    };

    fetchLogs();
  }, []);

  return (
    <div className="log-container">
      {logs}
    </div>
  );
};

export default Log;