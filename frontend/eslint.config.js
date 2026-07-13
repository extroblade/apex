import js from '@eslint/js';
import tseslint from 'typescript-eslint';
import reactHooks from 'eslint-plugin-react-hooks';
import reactRefresh from 'eslint-plugin-react-refresh';
import prettier from 'eslint-config-prettier';
import globals from 'globals';

export default tseslint.config(
  {
    ignores: [
      'dist',
      'node_modules',
      'storybook-static',
      '.storybook',
      '*.config.js',
      '*.config.ts',
      '*.config.mjs',
    ],
  },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      ecmaVersion: 2022,
      sourceType: 'module',
      globals: { ...globals.browser },
    },
    plugins: {
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
    },
    rules: {
      ...reactHooks.configs.recommended.rules,
      'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],
      '@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_' }],
    },
  },
  // Test files may use node globals and the DOM.
  {
    files: ['**/*.test.{ts,tsx}', 'src/test/**'],
    languageOptions: { globals: { ...globals.node } },
  },
  // Build/tooling scripts run under Node.
  {
    files: ['scripts/**'],
    languageOptions: { globals: { ...globals.node } },
  },
  // shadcn/ui primitives conventionally co-export a component and its variants.
  {
    files: ['src/shared/ui/**'],
    rules: { 'react-refresh/only-export-components': 'off' },
  },
  // Storybook stories export a default meta object plus named story objects.
  {
    files: ['**/*.stories.{ts,tsx}'],
    rules: { 'react-refresh/only-export-components': 'off' },
  },
  prettier,
);
