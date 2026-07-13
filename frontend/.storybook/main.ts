import type { StorybookConfig } from '@storybook/react-vite';
import { mergeConfig } from 'vite';
import { resolve } from 'node:path';

const config: StorybookConfig = {
  stories: ['../src/**/*.stories.@(ts|tsx)'],
  addons: ['@storybook/addon-essentials'],
  framework: { name: '@storybook/react-vite', options: {} },
  async viteFinal(cfg) {
    // Teach Storybook's Vite about the `@` -> src alias used across the app.
    return mergeConfig(cfg, {
      resolve: { alias: { '@': resolve(__dirname, '../src') } },
    });
  },
};

export default config;
