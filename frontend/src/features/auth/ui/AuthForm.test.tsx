import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { renderWithProviders } from '@/test/render';
import { AuthForm } from './AuthForm';

const login = vi.fn();
const register = vi.fn();

vi.mock('../api/use-auth', () => ({
  useLogin: () => ({ mutate: login, isPending: false, error: null }),
  useRegister: () => ({ mutate: register, isPending: false, error: null }),
}));

vi.mock('wouter', () => ({ useLocation: () => ['/', vi.fn()] }));

describe('AuthForm', () => {
  beforeEach(() => {
    login.mockReset();
    register.mockReset();
  });

  it('blocks submit and shows zod validation errors for bad input', async () => {
    const user = userEvent.setup();
    renderWithProviders(<AuthForm />);

    await user.type(screen.getByLabelText('Email'), 'not-an-email');
    await user.type(screen.getByLabelText('Password'), 'short');
    await user.click(screen.getByRole('button', { name: 'Log in' }));

    expect(await screen.findByText('Enter a valid email address.')).toBeInTheDocument();
    expect(
      screen.getByText('Password must be at least 8 characters.'),
    ).toBeInTheDocument();
    expect(login).not.toHaveBeenCalled();
  });

  it('submits valid credentials', async () => {
    const user = userEvent.setup();
    renderWithProviders(<AuthForm />);

    await user.type(screen.getByLabelText('Email'), 'racer@example.com');
    await user.type(screen.getByLabelText('Password'), 'supersecret');
    await user.click(screen.getByRole('button', { name: 'Log in' }));

    await waitFor(() =>
      expect(login).toHaveBeenCalledWith(
        { email: 'racer@example.com', password: 'supersecret' },
        expect.anything(),
      ),
    );
  });
});
