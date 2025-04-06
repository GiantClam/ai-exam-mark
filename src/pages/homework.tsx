import React, { useState } from 'react';
import { Upload, Form, Select, InputNumber, Button, message, Card, Typography, Spin, Divider, Result, Tag, List } from 'antd';
import { InboxOutlined, CheckCircleOutlined } from '@ant-design/icons';
import type { UploadProps, UploadFile } from 'antd';

const { Dragger } = Upload;
const { Title, Paragraph, Text } = Typography;

interface Answer {
  questionNumber?: string;
  studentAnswer: string;
  isCorrect?: boolean;
  correctAnswer?: string;
  correctSteps?: string;
  explanation?: string;
  evaluation?: string;
}

interface HomeworkResult {
  answers?: Answer[];
  feedback?: string;
  overallScore?: string;
}

interface FormValues {
  type: string;
  layout: string;
  pagesPerStudent: number;
}

const Homework: React.FC = () => {
  const [form] = Form.useForm<FormValues>();
  const [fileList, setFileList] = useState<UploadFile[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [result, setResult] = useState<string | null>(null);
  const [displayResults, setDisplayResults] = useState<boolean>(false);

  const uploadProps: UploadProps = {
    name: 'file',
    multiple: false,
    fileList,
    beforeUpload: (file: File) => {
      const isImage = file.type.startsWith('image/');
      const isPDF = file.type === 'application/pdf';
      
      if (!isImage && !isPDF) {
        message.error('只支持图片或PDF文件！');
        return false;
      }

      const isLt100M = file.size / 1024 / 1024 < 100;
      if (!isLt100M) {
        message.error('文件大小不能超过100MB！');
        return false;
      }

      setFileList([file as unknown as UploadFile]);
      return false;
    },
    onChange(info: any) {
      setFileList(info.fileList.slice(-1));
    },
  };

  const handleSubmit = async (values: FormValues) => {
    try {
      // 重置状态
      setResult(null);
      setDisplayResults(false);
      setLoading(true);
      
      console.log('开始上传作业，表单数据:', values);
      const formData = new FormData();
      const file = fileList[0]?.originFileObj;
      
      if (!file) {
        console.error('未选择文件');
        message.error('请选择要上传的文件！');
        setLoading(false);
        return;
      }

      console.log('准备上传文件:', {
        name: file.name,
        type: file.type,
        size: file.size
      });

      formData.append('file', file);
      formData.append('type', values.type);
      formData.append('layout', values.layout);
      formData.append('pagesPerStudent', values.pagesPerStudent.toString());

      console.log('开始发送上传请求...');
      const response = await fetch('http://localhost:8280/api/upload/homework', {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        console.error('上传请求失败:', {
          status: response.status,
          statusText: response.statusText
        });
        throw new Error('上传失败');
      }

      const data = await response.json();
      console.log('上传成功，服务器响应:', data);
      
      if (data.results) {
        // 服务器直接返回了批改结果
        setResult(data.results);
        setDisplayResults(true);
        message.success('作业批改完成！');
      } else {
        message.error('未收到批改结果');
      }
    } catch (error: any) {
      console.error('发生错误:', {
        message: error.message,
        stack: error.stack,
        error: error
      });
      message.error(error.message || '操作失败，请重试！');
    } finally {
      setLoading(false);
    }
  };

  const resetForm = () => {
    form.resetFields();
    setFileList([]);
    setResult(null);
    setDisplayResults(false);
  };

  const renderResults = () => {
    if (!result) return null;
    
    try {
      // 预处理文本，尝试清理可能的Markdown或其他非JSON字符
      let cleanedResult = result;
      
      // 如果结果以```json开头，去掉Markdown格式
      if (typeof cleanedResult === 'string') {
        // 移除可能的Markdown代码块标记
        cleanedResult = cleanedResult.replace(/```json|```/g, '');
        // 移除星号和其他可能导致问题的Markdown标记
        cleanedResult = cleanedResult.replace(/\*/g, '');
        // 移除开头的空格和换行
        cleanedResult = cleanedResult.trim();
      }
      
      // 尝试解析JSON结果
      let parsedResult;
      try {
        parsedResult = JSON.parse(cleanedResult);
      } catch (jsonError) {
        console.error("JSON解析失败，尝试查找文本中的JSON部分:", jsonError);
        
        // 尝试在文本中找到JSON对象
        const jsonMatch = cleanedResult.match(/(\{[\s\S]*\}|\[[\s\S]*\])/);
        if (jsonMatch && jsonMatch[0]) {
          try {
            parsedResult = JSON.parse(jsonMatch[0]);
            console.log("从文本中提取JSON成功");
          } catch (e) {
            console.error("从文本中提取JSON也失败:", e);
            throw e; // 向外抛出错误，交由外层处理
          }
        } else {
          throw jsonError; // 没找到JSON，向外抛出原始错误
        }
      }
      
      // 判断是否为数组类型（多个学生的结果）
      if (Array.isArray(parsedResult)) {
        return (
          <div style={{ marginTop: 24 }}>
            <Result
              icon={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
              title="作业批改完成"
              subTitle="系统已完成多名学生的作业分析"
            />
            
            {parsedResult.map((studentResult, index) => {
              let answers = [];
              let feedback = '';
              let overallScore = '';
              
              try {
                // 尝试解析每个学生的结果
                const student = JSON.parse(studentResult);
                answers = student.answers || [];
                feedback = student.feedback || '';
                overallScore = student.overallScore || '';
                
                return (
                  <Card 
                    key={index} 
                    title={`学生 ${index + 1} 批改结果`} 
                    style={{ marginTop: 16 }}
                    extra={overallScore ? <Tag color="blue">得分: {overallScore}</Tag> : null}
                  >
                    {answers.length > 0 && (
                      <>
                        <Title level={4}>题目分析</Title>
                        <List
                          itemLayout="vertical"
                          dataSource={answers}
                          renderItem={(answer: Answer, i) => (
                            <List.Item>
                              <Card
                                type="inner"
                                title={`题目 ${answer.questionNumber || i + 1}`}
                                extra={
                                  answer.isCorrect !== undefined ? (
                                    <Tag color={answer.isCorrect ? "success" : "error"}>
                                      {answer.isCorrect ? "正确" : "错误"}
                                    </Tag>
                                  ) : null
                                }
                              >
                                <Paragraph><strong>学生答案:</strong> {answer.studentAnswer}</Paragraph>
                                {answer.correctAnswer && (
                                  <Paragraph><strong>正确答案:</strong> {answer.correctAnswer}</Paragraph>
                                )}
                                {answer.correctSteps && (
                                  <Paragraph><strong>正确步骤:</strong> {answer.correctSteps}</Paragraph>
                                )}
                                {answer.explanation && (
                                  <Paragraph>
                                    <strong>解释:</strong> {answer.explanation}
                                  </Paragraph>
                                )}
                                {answer.evaluation && (
                                  <Paragraph>
                                    <strong>评价:</strong> {answer.evaluation}
                                  </Paragraph>
                                )}
                              </Card>
                            </List.Item>
                          )}
                        />
                      </>
                    )}
                    
                    {feedback && (
                      <>
                        <Divider />
                        <Title level={4}>整体评价</Title>
                        <Paragraph>{feedback}</Paragraph>
                      </>
                    )}
                  </Card>
                );
              } catch (e) {
                // 如果解析失败，直接显示原始内容
                return (
                  <Card key={index} title={`学生 ${index + 1} 批改结果`} style={{ marginTop: 16 }}>
                    <pre style={{ whiteSpace: 'pre-wrap', wordWrap: 'break-word' }}>
                      {studentResult}
                    </pre>
                  </Card>
                );
              }
            })}
            
            <div style={{ marginTop: 24, textAlign: 'center' }}>
              <Button type="primary" onClick={resetForm}>
                批改新作业
              </Button>
            </div>
          </div>
        );
      } else {
        // 单个学生的结果
        const answers = parsedResult.answers || [];
        const feedback = parsedResult.feedback || '';
        const overallScore = parsedResult.overallScore || '';
        
        return (
          <div style={{ marginTop: 24 }}>
            <Result
              icon={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
              title="作业批改完成"
              subTitle="系统已完成作业分析，以下是批改结果"
            />
            
            <Card
              style={{ marginTop: 16 }}
              title="批改结果"
              extra={overallScore ? <Tag color="blue">得分: {overallScore}</Tag> : null}
            >
              {answers.length > 0 && (
                <>
                  <Title level={4}>题目分析</Title>
                  <List
                    itemLayout="vertical"
                    dataSource={answers}
                    renderItem={(answer: Answer, index) => (
                      <List.Item>
                        <Card
                          type="inner"
                          title={`题目 ${answer.questionNumber || index + 1}`}
                          extra={
                            answer.isCorrect !== undefined ? (
                              <Tag color={answer.isCorrect ? "success" : "error"}>
                                {answer.isCorrect ? "正确" : "错误"}
                              </Tag>
                            ) : null
                          }
                        >
                          <Paragraph><strong>学生答案:</strong> {answer.studentAnswer}</Paragraph>
                          {answer.correctAnswer && (
                            <Paragraph><strong>正确答案:</strong> {answer.correctAnswer}</Paragraph>
                          )}
                          {answer.correctSteps && (
                            <Paragraph><strong>正确步骤:</strong> {answer.correctSteps}</Paragraph>
                          )}
                          {answer.explanation && (
                            <Paragraph>
                              <strong>解释:</strong> {answer.explanation}
                            </Paragraph>
                          )}
                          {answer.evaluation && (
                            <Paragraph>
                              <strong>评价:</strong> {answer.evaluation}
                            </Paragraph>
                          )}
                        </Card>
                      </List.Item>
                    )}
                  />
                </>
              )}
              
              {feedback && (
                <>
                  <Divider />
                  <Title level={4}>整体评价</Title>
                  <Paragraph>{feedback}</Paragraph>
                </>
              )}
            </Card>
            
            <div style={{ marginTop: 24, textAlign: 'center' }}>
              <Button type="primary" onClick={resetForm}>
                批改新作业
              </Button>
            </div>
          </div>
        );
      }
    } catch (e) {
      // 如果不是有效的JSON，尝试展示格式化的文本
      console.error("解析结果失败:", e);
      
      // 判断是否为纯文本响应
      let displayText = result;
      // 尝试进行一些基本格式化
      if (typeof displayText === 'string') {
        // 移除Markdown代码块标记
        displayText = displayText.replace(/```json|```/g, '');
        // 替换常见的Markdown标题格式为HTML
        displayText = displayText.replace(/#+\s+(.*?)(\n|$)/g, '<h3>$1</h3>');
        // 替换星号为HTML加粗
        displayText = displayText.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');
        displayText = displayText.replace(/\*(.*?)\*/g, '<em>$1</em>');
      }
      
      return (
        <div style={{ marginTop: 24 }}>
          <Result
            icon={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
            title="作业批改完成"
            subTitle="系统已完成作业分析，以下是批改结果"
          />
          
          <Card
            style={{ marginTop: 16 }}
            title="批改结果"
            extra={
              <Button
                type="link"
                onClick={() => message.info('尝试刷新页面或重新上传作业')}
              >
                解析格式有问题
              </Button>
            }
          >
            {typeof displayText === 'string' && displayText.includes('<h3>') ? (
              <div 
                style={{ padding: '10px', maxHeight: '500px', overflow: 'auto' }}
                dangerouslySetInnerHTML={{ __html: displayText }}
              />
            ) : (
              <pre style={{ whiteSpace: 'pre-wrap', wordWrap: 'break-word', maxHeight: '500px', overflow: 'auto' }}>
                {displayText}
              </pre>
            )}
          </Card>
          
          <div style={{ marginTop: 24, textAlign: 'center' }}>
            <Button type="primary" onClick={resetForm}>
              批改新作业
            </Button>
          </div>
        </div>
      );
    }
  };

  return (
    <div style={{ maxWidth: 800, margin: '0 auto', padding: '24px' }}>
      <Card>
        <Title level={2}>作业上传与批改</Title>
        <Paragraph>
          支持上传图片或PDF格式的作业文件。图片作业支持单栏和双栏布局，PDF作业支持按页数分割。
        </Paragraph>

        {!displayResults ? (
          <Form
            form={form}
            layout="vertical"
            onFinish={handleSubmit}
            style={{ marginTop: 24 }}
          >
            <Form.Item
              label="上传文件"
              required
            >
              <Dragger {...uploadProps}>
                <p className="ant-upload-drag-icon">
                  <InboxOutlined />
                </p>
                <p className="ant-upload-text">点击或拖拽文件到此区域上传</p>
                <p className="ant-upload-hint">
                  支持单个图片或PDF文件，大小不超过100MB
                </p>
              </Dragger>
            </Form.Item>

            <Form.Item
              label="作业类型"
              name="type"
              initialValue="general"
              rules={[{ required: true, message: '请选择作业类型' }]}
            >
              <Select>
                <Select.Option value="general">通用作业</Select.Option>
                <Select.Option value="math">数学作业</Select.Option>
                <Select.Option value="english">英语作业</Select.Option>
                <Select.Option value="chinese">语文作业</Select.Option>
              </Select>
            </Form.Item>

            <Form.Item
              label="布局方式"
              name="layout"
              initialValue="single"
              rules={[{ required: true, message: '请选择布局方式' }]}
            >
              <Select>
                <Select.Option value="single">单栏布局</Select.Option>
                <Select.Option value="double">双栏布局</Select.Option>
              </Select>
            </Form.Item>

            <Form.Item
              label="每个学生作业页数"
              name="pagesPerStudent"
              initialValue={1}
              rules={[{ required: true, message: '请输入每个学生的作业页数' }]}
            >
              <InputNumber min={1} max={10} />
            </Form.Item>

            <Form.Item>
              <Button type="primary" htmlType="submit" loading={loading} disabled={loading}>
                {loading ? '批改中...' : '上传并开始批改'}
              </Button>
            </Form.Item>
          </Form>
        ) : (
          renderResults()
        )}

        {loading && (
          <div style={{ textAlign: 'center', marginTop: 24 }}>
            <Spin size="large" tip="正在批改作业中，请稍候..." />
            <Paragraph style={{ marginTop: 16 }}>
              作业批改可能需要几分钟时间，取决于文件大小和复杂度
            </Paragraph>
          </div>
        )}
      </Card>
    </div>
  );
};

export default Homework; 