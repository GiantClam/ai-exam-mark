import React, { useState, useEffect } from 'react';
import { Card, Typography, Spin, message, Button } from 'antd';
import { useRouter } from 'next/router';
import { ReloadOutlined } from '@ant-design/icons';

const { Title, Paragraph } = Typography;

const MarkHomework: React.FC = () => {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    console.log('MarkHomework 组件加载，路由参数:', router.query);
    
    if (router.query.result) {
      console.log('收到批改结果:', router.query.result);
      setResult(router.query.result as string);
      setError(null);
    } else {
      console.warn('未收到批改结果');
      setError('未找到批改结果，请重新上传作业');
    }
  }, [router.query.result]);

  const handleRetry = () => {
    console.log('用户点击重试按钮');
    router.back();
  };

  return (
    <div style={{ maxWidth: 800, margin: '0 auto', padding: '24px' }}>
      <Card>
        <Title level={2}>作业批改结果</Title>
        <Paragraph>
          系统已完成作业分析，以下是批改结果：
        </Paragraph>

        {loading && (
          <div style={{ textAlign: 'center', marginTop: 24 }}>
            <Spin size="large" tip="正在批改中..." />
          </div>
        )}

        {error && (
          <div style={{ marginTop: 24, textAlign: 'center' }}>
            <Paragraph type="danger">{error}</Paragraph>
            <Button 
              type="primary" 
              icon={<ReloadOutlined />} 
              onClick={handleRetry}
            >
              重试
            </Button>
          </div>
        )}

        {result && (
          <div style={{ marginTop: 24 }}>
            <Title level={3}>批改结果</Title>
            <Card>
              <pre style={{ whiteSpace: 'pre-wrap', wordWrap: 'break-word' }}>
                {result}
              </pre>
            </Card>
          </div>
        )}
      </Card>
    </div>
  );
};

export default MarkHomework; 