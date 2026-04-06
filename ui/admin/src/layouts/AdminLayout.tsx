import { Outlet, Navigate } from 'react-router-dom';
import Sidebar from '../app/components/Sidebar';
import { useAuthStore } from '../store/authStore';

export default function AdminLayout() {
  const { token } = useAuthStore();

  if (!token) {
    return <Navigate to="/login" replace />;
  }

  return (
    <div className="flex h-screen bg-gray-50">
      <Sidebar />
      <main className="flex-1 overflow-y-auto p-8">
        <Outlet />
      </main>
    </div>
  );
}
