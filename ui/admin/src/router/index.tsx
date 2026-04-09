import { createBrowserRouter, Navigate } from 'react-router-dom';
import AdminLayout from '../layouts/AdminLayout';
import AuthLayout from '../layouts/AuthLayout';
import ArticleSectionLayout from '../layouts/ArticleSectionLayout';
import Login from '../pages/login';
import Dashboard from '../pages/dashboard';
import ArticleManagement from '../app/components/ArticleManagement';
import ArticleEdit from '../pages/article/edit';
import ArticleTrash from '../pages/article/trash';
import Categories from '../pages/category';
import Tags from '../pages/tag';
import Comments from '../pages/comment';
import Attachments from '../pages/attachment';
import Settings from '../pages/setting';
import Users from '../pages/user';
import PageManagement from '../app/components/PageManagement';
import PageEdit from '../pages/page/edit';
import PageTrash from '../pages/page/trash';

const router = createBrowserRouter([
  {
    path: '/login',
    element: <AuthLayout />,
    children: [
      {
        index: true,
        element: <Login />,
      },
    ],
  },
  {
    path: '/',
    element: <AdminLayout />,
    children: [
      {
        index: true,
        element: <Dashboard />,
      },
      {
        path: 'articles/create',
        element: <ArticleEdit />,
      },
      {
        path: 'articles/edit/:id',
        element: <ArticleEdit />,
      },
      {
        path: 'articles',
        element: <ArticleSectionLayout />,
        children: [
          {
            index: true,
            element: <ArticleManagement />,
          },
          {
            path: 'categories',
            element: <Categories />,
          },
          {
            path: 'tags',
            element: <Tags />,
          },
          {
            path: 'trash',
            element: <ArticleTrash />,
          },
        ],
      },
      {
        path: 'categories',
        element: <Navigate to="/articles/categories" replace />,
      },
      {
        path: 'tags',
        element: <Navigate to="/articles/tags" replace />,
      },
      {
        path: 'pages/create',
        element: <PageEdit />,
      },
      {
        path: 'pages/edit/:id',
        element: <PageEdit />,
      },
      {
        path: 'pages/trash',
        element: <PageTrash />,
      },
      {
        path: 'pages',
        element: <PageManagement />,
      },
      {
        path: 'comments',
        element: <Comments />,
      },
      {
        path: 'attachments',
        element: <Attachments />,
      },
      {
        path: 'settings',
        element: <Settings />,
      },
      {
        path: 'users',
        element: <Users />,
      },
    ],
  },
]);

export default router;
