import React, {useState} from 'react'
import type { DataNode } from 'antd/es/tree';
import {ProFormCheckbox} from "@ant-design/pro-components";
import { Tree } from 'antd';

type Props = {
  treeData:DataNode[]
  value?: number[];
  onChange?: (value: number[]) => void;
}


interface CheckInfo {
  event: 'check';
  node: DataNode;
  checked: boolean;
  checkedNodes: DataNode[];
}

const MenuTree:React.FC<Props> = props => {
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>();
  const [checkedKeys, setCheckedKeys] = useState<React.Key[]>(props.value || []);

  // const [selectedKeys, setSelectedKeys] = useState<React.Key[]>([]);
  // const [autoExpandParent, setAutoExpandParent] = useState<boolean>(true);
  const triggerChange = (changedValue: number[]) => {
    props.onChange?.(changedValue);
  };
  const onExpand = (expandedKeysValue: React.Key[]) => {
    setExpandedKeys(expandedKeysValue);
  };

  const handleCheck = (
    checked: React.Key[] | { checked: React.Key[]; halfChecked: React.Key[] },
    info: CheckInfo
  ): void => {
    if (Array.isArray(checked)) {
      // 当checked是Key[]类型时的处理逻辑
      console.log('Checked keys:', props.value);
      let select: number[] = [];
      checked.forEach((item) => {
        select.push(Number(item));
      });
      triggerChange(select);
      setCheckedKeys(checked);
    } else {
      // 当checked是对象{ checked: Key[]; halfChecked: Key[] }时的处理逻辑
      console.log('Checked keys:', checked.checked);
      console.log('Half-checked keys:', checked.halfChecked);
    }
    // 这里可以使用info做一些额外的处理
    console.log('Check info:', info);
  };

  function keyToNumber(key: React.Key): number  {
    if (typeof key === 'number') {
      return key;
    } else if (typeof key === 'string') {
      // 尝试将字符串转换为数字
      const parsed = Number(key);
      // 检查转换结果是否为有效数字
      return isNaN(parsed) ? 0 : parsed;
    }
    // 如果 key 类型不是 string 或 number，则返回 null
    // 在 React.Key 的上下文中，这种情况不应该发生
    return 0;
  }
  const getAllKeys = (nodes: DataNode[]): number[] => {
    let keys: number[] = [];
    nodes.forEach(node => {
      keys.push(keyToNumber(node.key)); // 收集当前节点的key
      if (node.children && node.children.length) {
        keys = keys.concat(getAllKeys(node.children)); // 递归收集子节点的keys
      }
    });
    return keys;
  };


  const onSelect = (selectedKeysValue: React.Key[], info: any) => {
    console.log('onSelect', info);
    // setSelectedKeys(selectedKeysValue);
    let select: number[] = [];
    selectedKeysValue.forEach((item) => {
      select.push(Number(item));
    });
    triggerChange(select);

  };

  return (
    <>
      <ProFormCheckbox.Group
        fieldProps={{
          onChange: (values) => {
            setExpandedKeys(values.includes("展开/折叠") ? getAllKeys(props.treeData) : []);
            setCheckedKeys(values.includes("全选/全不选") ? getAllKeys(props.treeData) : []);
          }}}
        name="checkbox"
        layout="horizontal"
        label="菜单权限"
        options={['展开/折叠', '全选/全不选',]}
      />
      <Tree
        style={{border: '1px solid #eee'}}
        checkable
        blockNode
        checkedKeys={checkedKeys}
        onExpand={onExpand}
        expandedKeys={expandedKeys}
        // autoExpandParent={autoExpandParent}
        onCheck={handleCheck}

        onSelect={onSelect}
        selectedKeys={props.value}
        treeData={props.treeData}
      />
    </>
  )
}

export default MenuTree
