// eslint.config.js
import globals from 'globals';
import pluginJs from '@eslint/js';
import tseslint from 'typescript-eslint';
import pluginReactConfig from 'eslint-plugin-react/configs/recommended.js';
import pluginReactHooks from 'eslint-plugin-react-hooks';
import pluginJsxA11y from 'eslint-plugin-jsx-a11y';
import pluginPrettierRecommended from 'eslint-plugin-prettier/recommended';
import pluginReactRefresh from 'eslint-plugin-react-refresh';

export default [
  { files: ['**/*.{js,mjs,cjs,ts,jsx,tsx}'] },
  {
    languageOptions: {
      globals: { ...globals.browser, ...globals.node, ...globals.es2020 },
      parser: tseslint.parser, // 指定TS解析器
      parserOptions: {
        project: ['./tsconfig.json', './tsconfig.node.json'], // 需要tsconfig来启用类型相关的linting规则
        ecmaVersion: 'latest',
        sourceType: 'module',
        ecmaFeatures: {
          jsx: true,
        },
      },
    },
  },

  pluginJs.configs.recommended,

  ...tseslint.configs.recommendedTypeChecked, // 使用类型检查的规则集

  {
    files: ['**/*.{jsx,tsx}'],
    ...pluginReactConfig,
    settings: {
      react: {
        version: 'detect',
      },
    },
    rules: {
      ...pluginReactConfig.rules,
      'react/react-in-jsx-scope': 'off',
      'react/jsx-uses-react': 'off',
      'react/prop-types': 'off',
    },
  },

  {
    plugins: {
      'react-hooks': pluginReactHooks,
    },
    rules: pluginReactHooks.configs.recommended.rules,
  },

  {
    plugins: {
      'jsx-a11y': pluginJsxA11y,
    },
    // eslint-plugin-jsx-a11y 的 configs.recommended 不直接是规则对象
    // 需要手动查看其推荐规则或查找其在扁平配置中的用法
    // 暂时留空或参考其文档
  },

  {
    plugins: {
      'react-refresh': pluginReactRefresh,
    },
    rules: {
      'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],
    },
  },

  pluginPrettierRecommended,

  {
    rules: {
      '@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_' }],
    },
  },

  {
    ignores: ['dist/', 'eslint.config.js', 'node_modules/'],
  },
];
