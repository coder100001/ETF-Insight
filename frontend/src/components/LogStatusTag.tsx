import React from 'react';
import { Tag } from 'antd';
import styled from 'styled-components';

interface LogStatusTagProps {
  status: 'success' | 'failure' | 'warning' | 'info';
  children?: React.ReactNode;
}

const StyledTag = styled(Tag)`
  margin: 0;
  font-weight: 500;
  border-radius: 4px;
`;

const getStatusColor = (status: LogStatusTagProps['status']): string => {
  switch (status) {
    case 'success':
      return '#2ecc71';
    case 'failure':
      return '#e74c3c';
    case 'warning':
      return '#f39c12';
    case 'info':
      return '#3498db';
    default:
      return '#95a5a6';
  }
};

const LogStatusTag: React.FC<LogStatusTagProps> = ({ status, children }) => {
  const color = getStatusColor(status);
  const text = children || (status === 'success' ? '成功' : status === 'failure' ? '失败' : status);

  return (
    <StyledTag
      color={color}
      style={{
        color: '#fff',
        border: 'none',
        padding: '2px 8px',
        fontSize: '12px',
      }}
    >
      {text}
    </StyledTag>
  );
};

export default LogStatusTag;
