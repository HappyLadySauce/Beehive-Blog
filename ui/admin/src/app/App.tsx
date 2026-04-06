import { useState } from 'react';
import Sidebar from './components/Sidebar';
import Dashboard from './components/Dashboard';
import ArticleManagement from './components/ArticleManagement';

export default function App() {
  const [activeTab, setActiveTab] = useState('dashboard');

  const renderContent = () => {
    switch (activeTab) {
      case 'dashboard':
        return <Dashboard />;
      case 'articles':
        return <ArticleManagement />;
      case 'pages':
      case 'categories':
      case 'tags':
      case 'comments':
      case 'attachments':
      case 'users':
      case 'settings':
      case 'quickcreate':
        return (
          <div className="bg-white border border-gray-200 rounded p-8 text-center">
            <div className="text-gray-400 text-lg mb-2">功能开发中...</div>
            <div className="text-gray-500 text-sm">该功能即将上线</div>
          </div>
        );
      default:
        return <Dashboard />;
    }
  };

  return (
    <div className="size-full flex bg-gray-50">
      <Sidebar activeTab={activeTab} onTabChange={setActiveTab} />
      <main className="flex-1 overflow-auto">
        <div className="p-6">
          {renderContent()}
        </div>
      </main>
    </div>
  );
}