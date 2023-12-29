import React from 'react';
import { Space, Input, Select } from 'antd';

interface API {
  method?: string;
  url?: string;
}

type Props = {
  value?: API;
  onChange?: (value: API) => void;
};

const options = [
  {
    value: 'GET',
    label: 'GET',
  },
  {
    value: 'PUT',
    label: 'PUT',
  },
  {
    value: 'POST',
    label: 'POST',
  },
  {
    value: 'DELETE',
    label: 'DELETE',
  },
];
const APIInput: React.FC<Props> = (props) => {
  let api: API = {
    method: 'GET',
    url: '',
  };

  return (
    <div>
      {/*<p>*/}
      {/*  <span style={{ color: 'red' }}>＊</span> 接口规则*/}
      {/*</p>*/}
      <Space.Compact style={{ width: '328px' }}>
        <Select
          style={{ width: '120px' }}
          defaultValue={
            props.value?.method === undefined || props.value?.method === ''
              ? 'GET'
              : props.value?.method
          }
          options={options}
          onChange={(value) => {
            api.method = value;
            props.onChange?.(api);
          }}
        />
        <Input
          placeholder="后端api地址"
          defaultValue={props.value?.url}
          onChange={(e) => {
            api.url = e.target.value;
            props.onChange?.(api);
          }}
        />
      </Space.Compact>
    </div>
  );
};

export default APIInput;
