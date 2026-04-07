import React from 'react';
import ReactDOM from 'react-dom/client';
import { ThemeProvider } from 'next-themes';
import { RouterProvider } from 'react-router-dom';
import router from './router';
import './styles/index.css';
import './styles/tailwind.css';
import './styles/theme.css';
import { Toaster } from './app/components/ui/sonner';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ThemeProvider
      attribute="class"
      defaultTheme="system"
      enableSystem
      storageKey="beehive-admin-theme"
    >
      <RouterProvider router={router} />
      <Toaster />
    </ThemeProvider>
  </React.StrictMode>,
);
