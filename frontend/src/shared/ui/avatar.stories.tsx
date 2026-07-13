import type { Meta, StoryObj } from '@storybook/react';

import { Avatar, AvatarFallback, AvatarImage } from './avatar';

const meta: Meta<typeof Avatar> = {
  title: 'UI/Avatar',
  component: Avatar,
};
export default meta;

type Story = StoryObj<typeof Avatar>;

export const Fallback: Story = {
  render: () => (
    <Avatar>
      <AvatarFallback>AL</AvatarFallback>
    </Avatar>
  ),
};

export const WithImage: Story = {
  render: () => (
    <Avatar className="size-16">
      <AvatarImage src="https://i.pravatar.cc/128" alt="" />
      <AvatarFallback>AL</AvatarFallback>
    </Avatar>
  ),
};
