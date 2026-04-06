import { createBrowserRouter } from 'react-router-dom';
import AdminLayout from '../layouts/AdminLayout';
import AuthLayout from '../layouts/AuthLayout';
import Login from '../pages/login';
import Dashboard from '../pages/dashboard';
import Articles from '../pages/article';
import ArticleEdit from '../pages/article/edit';
import Categories from '../pages/category';
import Tags from '../pages/tag';
import Comments from '../pages/comment';
import Attachments from '../pages/attachment';
import Settings from '../pages/setting';
import Users from '../pages/user';

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
        path: 'articles',
        element: <Articles />,
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
        path: 'categories',
        element: <Categories />,
      },
      {
        path: 'tags',
        element: <Tags />,
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
