import React, { useCallback, useMemo, useState, useEffect } from 'react';
import { Input, Tree } from 'antd';
import type { DataNode } from 'antd/es/tree';
import { getDepartments } from '@/services/echo-admin/bumen';

const { Search } = Input;

// 将部门数据转换为树形结构的数据
const departmentItemTree = (menus: API.department[]): DataNode[] =>
  menus.map(({ name, id, children }) => ({
    title: name,
    key: id ?? 0,
    children: children && children.length > 0 ? departmentItemTree(children) : [],
  }));

// 获取父节点的 key
const getParentKey = (key: React.Key, tree: DataNode[]): React.Key | undefined => {
  for (const node of tree) {
    if (node.children?.some((item) => item.key === key)) {
      return node.key;
    }
    const parentKey = node.children && getParentKey(key, node.children);
    if (parentKey) return parentKey;
  }
};

export type TreeProps = {
  onSelect: (values: any) => Promise<void>;
};
const Department: React.FC<TreeProps> = (props) => {
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);
  const [searchValue, setSearchValue] = useState('');
  const [autoExpandParent, setAutoExpandParent] = useState(true);
  const [defaultData, setDefaultData] = useState<DataNode[]>([]);

  // 在组件挂载时获取部门数据
  useEffect(() => {
    const fetchData = async () => {
      try {
        const result = await getDepartments({ page: 1, count: 1000 });
        setDefaultData(departmentItemTree(result.data.list ?? []));
      } catch (error) {
        console.error('获取数据出错:', error);
      }
    };

    fetchData();
  }, []);

  // 展开或收起树节点时的回调
  const onExpand = useCallback((newExpandedKeys: React.Key[]) => {
    setExpandedKeys(newExpandedKeys);
    setAutoExpandParent(false);
  }, []);

  // 搜索框内容变化时的回调
  const onChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const { value } = e.target;
      setSearchValue(value); // 更新搜索值

      if (!value) {
        setExpandedKeys([]); // 如果搜索值为空，重置展开的键
      } else {
        // 如果有搜索值，找出所有匹配项的键和它们的父键
        const expandedKeys = defaultData
          .reduce((accumulator, item) => {
            const matches = item.title?.toString().toLowerCase().includes(value.toLowerCase());
            let newAccumulator = [...accumulator];

            if (matches) {
              newAccumulator.push(item.key);
              // 如果当前项匹配，也需要将其父项的 key 添加到展开数组中
              const parentKey = getParentKey(item.key, defaultData);
              if (parentKey) {
                newAccumulator.push(parentKey);
              }
            }
            if (item.children) {
              // 递归检查子项
              const childKeys = item.children
                .filter(
                  (child) => child.title?.toString().toLowerCase().includes(value.toLowerCase()),
                )
                .map((child) => child.key);
              newAccumulator = [...newAccumulator, ...childKeys];
            }
            return newAccumulator;
          }, [] as React.Key[])
          .filter((key, index, self) => self.indexOf(key) === index); // 去重

        setExpandedKeys(expandedKeys);
      }
      setAutoExpandParent(true); // 允许自动展开父项
    },
    [defaultData],
  );

  const onSelect = (keys: React.Key[], info: any) => {
    console.log(keys);
    console.log(info);
    if (keys.length > 0) {
      props.onSelect(keys[0]);
    }
  };
  // 使用 useMemo 来优化树节点的渲染
  const treeData = useMemo(() => {
    const loop = (data: DataNode[]): DataNode[] =>
      data.map((item) => {
        const strTitle = item.title as string;
        const index = item.title?.toString().indexOf(searchValue) ?? -1;
        const beforeStr = item.title?.toString().substring(0, index);
        const afterStr = item.title?.toString().substring(index + searchValue.length);
        const title =
          index > -1 ? (
            <span>
              {beforeStr}
              <span className="site-tree-search-value">{searchValue}</span>
              {afterStr}
            </span>
          ) : (
            <span>{strTitle}</span>
          );

        return {
          title,
          key: item.key,
          children: item.children ? loop(item.children) : undefined,
        };
      });

    return loop(defaultData);
  }, [searchValue, defaultData]);

  return (
    <div>
      <Search style={{ marginBottom: 8 }} placeholder="搜索" onChange={onChange} />
      <Tree
        onExpand={onExpand}
        expandedKeys={expandedKeys}
        autoExpandParent={autoExpandParent}
        treeData={treeData}
        onSelect={onSelect}
      />
    </div>
  );
};

export default Department;
