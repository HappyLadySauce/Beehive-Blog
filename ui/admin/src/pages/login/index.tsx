import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../../store/authStore';
import { login, getUserInfo } from '../../api/auth';
import { Button } from '../../app/components/ui/button';
import { Input } from '../../app/components/ui/input';
import { Label } from '../../app/components/ui/label';
import { toast } from 'sonner';

export default function Login() {
  const [account, setAccount] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { setTokens, setUser } = useAuthStore();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!account || !password) {
      toast.error('请输入账号和密码');
      return;
    }

    setLoading(true);
    try {
      const res = await login({ account, password });
      if (res.code === 200) {
        setTokens(res.data.token, res.data.refreshToken);
        
        // Fetch user info
        const userRes = await getUserInfo();
        if (userRes.code === 200) {
          setUser(userRes.data);
          toast.success('登录成功');
          navigate('/');
        } else {
          toast.error(userRes.message || '获取用户信息失败');
        }
      } else {
        toast.error(res.message || '登录失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '登录请求失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-card border border-border py-8 px-4 shadow sm:rounded-lg sm:px-10">
      <div className="sm:mx-auto sm:w-full sm:max-w-md mb-6">
        <div className="flex justify-center">
          <div className="w-12 h-12 bg-primary rounded flex items-center justify-center text-primary-foreground text-xl font-bold">
            B
          </div>
        </div>
        <h2 className="mt-6 text-center text-3xl font-extrabold text-foreground">
          Beehive Blog Admin
        </h2>
      </div>

      <form className="space-y-6" onSubmit={handleLogin}>
        <div>
          <Label htmlFor="account">账号 / 邮箱</Label>
          <div className="mt-1">
            <Input
              id="account"
              name="account"
              type="text"
              autoComplete="username"
              required
              value={account}
              onChange={(e) => setAccount(e.target.value)}
              placeholder="请输入用户名或邮箱"
            />
          </div>
        </div>

        <div>
          <Label htmlFor="password">密码</Label>
          <div className="mt-1">
            <Input
              id="password"
              name="password"
              type="password"
              autoComplete="current-password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="请输入密码"
            />
          </div>
        </div>

        <div>
          <Button
            type="submit"
            className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-primary-foreground bg-primary hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-ring"
            disabled={loading}
          >
            {loading ? '登录中...' : '登录'}
          </Button>
        </div>
      </form>
    </div>
  );
}
