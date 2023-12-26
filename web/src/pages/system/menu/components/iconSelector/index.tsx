import * as React from 'react';
import Icon, * as AntdIcons from '@ant-design/icons';
import { Radio, Input, Empty, Flex } from 'antd';
import type { RadioChangeEvent } from 'antd/es/radio/interface';
import debounce from 'lodash/debounce';
import Category from './category';
import { FilledIcon, OutlinedIcon, TwoToneIcon } from './themeIcons';
import type { CategoriesKeys } from './fields';
import { categories } from './fields';

export enum ThemeType {
  Filled = 'Filled',
  Outlined = 'Outlined',
  TwoTone = 'TwoTone',
}

const allIcons: {
  [key: string]: any;
} = AntdIcons;

interface IconSelectorProps {
  onSelect: (iconName: string, iconComponent: typeof Icon) => void;
}

interface IconSelectorState {
  theme: ThemeType;
  searchKey: string;
}

const Index: React.FC<IconSelectorProps> = (props) => {
  const { onSelect } = props;
  const [displayState, setDisplayState] = React.useState<IconSelectorState>({
    theme: ThemeType.Outlined,
    searchKey: '',
  });

  const newIconNames: string[] = [];

  const handleSearchIcon = React.useCallback(
    debounce((searchKey: string) => {
      setDisplayState((prevState) => ({ ...prevState, searchKey }));
    }),
    [],
  );

  const handleChangeTheme = React.useCallback((e: RadioChangeEvent) => {
    setDisplayState((prevState) => ({ ...prevState, theme: e.target.value as ThemeType }));
  }, []);

  const renderCategories = React.useMemo<React.ReactNode | React.ReactNode[]>(() => {
    const { searchKey = '', theme } = displayState;

    const categoriesResult = (Object.keys(categories) as CategoriesKeys[])
      .map((key: CategoriesKeys) => {
        let iconList = categories[key];
        if (searchKey) {
          const matchKey = searchKey
            // eslint-disable-next-line prefer-regex-literals
            .replace(new RegExp(`^<([a-zA-Z]*)\\s/>$`, 'gi'), (_, name) => name)
            .replace(/(Filled|Outlined|TwoTone)$/, '')
            .toLowerCase();
          iconList = iconList.filter((iconName: string) =>
            iconName.toLowerCase().includes(matchKey),
          );
        }

        // CopyrightCircle is same as Copyright, don't show it
        iconList = iconList.filter((icon: string) => icon !== 'CopyrightCircle');

        return {
          category: key,
          icons: iconList
            .map((iconName: string) => iconName + theme)
            .filter((iconName: string) => allIcons[iconName]),
        };
      })
      .filter(({ icons }) => !!icons.length)
      .map(({ category, icons }) => (
        <Category
          key={category}
          title={category as CategoriesKeys}
          theme={theme}
          icons={icons}
          newIcons={newIconNames}
          onSelect={(type, name) => {
            if (onSelect) {
              onSelect(name, allIcons[name]);
            }
          }}
        />
      ));
    return categoriesResult.length === 0 ? <Empty style={{ margin: '2em 0' }} /> : categoriesResult;
  }, [displayState.searchKey, displayState.theme]);
  return (
    <>
      <Flex gap="middle" align="start" vertical>
        <Radio.Group
          value={displayState.theme}
          onChange={handleChangeTheme}
          size="large"
          optionType="button"
          buttonStyle="solid"
          options={[
            {
              label: <Icon component={OutlinedIcon} />,
              value: ThemeType.Outlined,
            },
            {
              label: <Icon component={FilledIcon} />,
              value: ThemeType.Filled,
            },
            {
              label: <Icon component={TwoToneIcon} />,
              value: ThemeType.TwoTone,
            },
          ]}
        ></Radio.Group>
        <Input.Search
          style={{ margin: '0 10px', flex: 1 }}
          allowClear
          onChange={(e) => handleSearchIcon(e.currentTarget.value)}
          size="large"
          autoFocus
        />
        {renderCategories}
      </Flex>
    </>
  );
};

export default Index;
