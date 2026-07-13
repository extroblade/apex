import type { Meta, StoryObj } from '@storybook/react';

import { ThemeToggle } from './ThemeToggle';

const meta: Meta<typeof ThemeToggle> = {
  title: 'Theme/ThemeToggle',
  component: ThemeToggle,
};
export default meta;

export const Default: StoryObj<typeof ThemeToggle> = {};
