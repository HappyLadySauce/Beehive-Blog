import { useState, useEffect } from 'react';
import { getSettings, updateSettings, testSmtp, syncHexoPosts, getHexoSyncStatus } from '../../api/setting';
import { toast } from 'sonner';
import { Settings as SettingsIcon, RefreshCw, Mail, Save } from 'lucide-react';

const tabs = [
  { id: 'general', label: '基础设置' },
  { id: 'seo', label: 'SEO 设置' },
  { id: 'smtp', label: '邮件服务 (SMTP)' },
  { id: 'hexo', label: 'Hexo 同步' },
];

export default function Settings() {
  const [activeTab, setActiveTab] = useState('general');
  const [settings, setSettings] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [testEmail, setTestEmail] = useState('');
  const [syncStatus, setSyncStatus] = useState<any>(null);

  const fetchSettings = async (group: string) => {
    setLoading(true);
    try {
      const res = await getSettings(group);
      if (res.code === 200) {
        setSettings(res.data.settings || {});
      } else {
        toast.error(res.message || '获取设置失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '请求设置失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchSyncStatus = async () => {
    try {
      const res = await getHexoSyncStatus();
      if (res.code === 200) {
        setSyncStatus(res.data);
      }
    } catch {
      // 静默失败
    }
  };

  useEffect(() => {
    if (activeTab !== 'hexo') {
      fetchSettings(activeTab);
    } else {
      fetchSyncStatus();
    }
  }, [activeTab]);

  const handleSave = async () => {
    setSaving(true);
    try {
      const res = await updateSettings(activeTab, settings);
      if (res.code === 200) {
        toast.success('保存成功');
      } else {
        toast.error(res.message || '保存失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '保存请求失败');
    } finally {
      setSaving(false);
    }
  };

  const handleTestSmtp = async () => {
    if (!testEmail) {
      toast.error('请输入测试接收邮箱');
      return;
    }
    try {
      const res = await testSmtp(testEmail);
      if (res.code === 200) {
        toast.success('测试邮件发送成功');
      } else {
        toast.error(res.message || '测试邮件发送失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '测试请求失败');
    }
  };

  const handleSyncHexo = async () => {
    try {
      const res = await syncHexoPosts(true);
      if (res.code === 200) {
        toast.success('同步成功');
        fetchSyncStatus();
      } else {
        toast.error(res.message || '同步失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '同步请求失败');
    }
  };

  const handleChange = (key: string, value: string) => {
    setSettings(prev => ({ ...prev, [key]: value }));
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <SettingsIcon className="w-5 h-5 text-muted-foreground" />
        <h2 className="text-lg font-medium text-foreground">系统设置</h2>
      </div>

      <div className="flex flex-col md:flex-row gap-6">
        <div className="w-full md:w-48 flex-shrink-0">
          <nav className="flex flex-col space-y-1">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`px-3 py-2 text-sm font-medium rounded-md text-left transition-colors ${
                  activeTab === tab.id
                    ? 'bg-primary/10 text-primary'
                    : 'text-foreground hover:bg-accent'
                }`}
              >
                {tab.label}
              </button>
            ))}
          </nav>
        </div>

        <div className="flex-1 bg-card border border-border rounded-lg p-6">
          {loading ? (
            <div className="py-12 text-center text-muted-foreground">加载中...</div>
          ) : (
            <div className="space-y-6 max-w-2xl">
              {activeTab === 'general' && (
                <>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">站点名称</label>
                    <input
                      type="text"
                      value={settings['site_name'] || ''}
                      onChange={(e) => handleChange('site_name', e.target.value)}
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">站点 URL</label>
                    <input
                      type="url"
                      value={settings['site_url'] || ''}
                      onChange={(e) => handleChange('site_url', e.target.value)}
                      placeholder="https://example.com"
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">站点描述</label>
                    <textarea
                      value={settings['site_description'] || ''}
                      onChange={(e) => handleChange('site_description', e.target.value)}
                      rows={3}
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                    />
                  </div>
                </>
              )}

              {activeTab === 'seo' && (
                <>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">SEO 标题</label>
                    <input
                      type="text"
                      value={settings['seo_title'] || ''}
                      onChange={(e) => handleChange('seo_title', e.target.value)}
                      placeholder="默认使用站点名称"
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">SEO 描述</label>
                    <textarea
                      value={settings['seo_description'] || ''}
                      onChange={(e) => handleChange('seo_description', e.target.value)}
                      rows={3}
                      placeholder="默认使用站点描述"
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">关键词</label>
                    <input
                      type="text"
                      value={settings['site_keywords'] || ''}
                      onChange={(e) => handleChange('site_keywords', e.target.value)}
                      placeholder="多个关键词用英文逗号分隔"
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">OG 图片 URL</label>
                    <input
                      type="url"
                      value={settings['site_og_image'] || ''}
                      onChange={(e) => handleChange('site_og_image', e.target.value)}
                      placeholder="https://example.com/og-image.png"
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                    />
                  </div>
                </>
              )}

              {activeTab === 'smtp' && (
                <>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">SMTP 服务器</label>
                    <input
                      type="text"
                      value={settings['smtp_host'] || ''}
                      onChange={(e) => handleChange('smtp_host', e.target.value)}
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">SMTP 端口</label>
                    <input
                      type="text"
                      value={settings['smtp_port'] || ''}
                      onChange={(e) => handleChange('smtp_port', e.target.value)}
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">发件人邮箱</label>
                    <input
                      type="email"
                      value={settings['smtp_user'] || ''}
                      onChange={(e) => handleChange('smtp_user', e.target.value)}
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">密码 / 授权码</label>
                    <input
                      type="password"
                      value={settings['smtp_pass'] || ''}
                      onChange={(e) => handleChange('smtp_pass', e.target.value)}
                      placeholder="留空表示不修改"
                      className="w-full px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                    />
                  </div>

                  <div className="pt-4 border-t border-border mt-6">
                    <h4 className="text-sm font-medium text-foreground mb-3">发送测试邮件</h4>
                    <div className="flex gap-3">
                      <input
                        type="email"
                        value={testEmail}
                        onChange={(e) => setTestEmail(e.target.value)}
                        placeholder="接收测试邮件的邮箱地址"
                        className="flex-1 px-3 py-2 border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent text-sm"
                      />
                      <button
                        onClick={handleTestSmtp}
                        className="px-4 py-2 bg-muted text-foreground rounded hover:bg-accent transition-colors flex items-center gap-2 text-sm font-medium"
                      >
                        <Mail className="w-4 h-4" />
                        测试发送
                      </button>
                    </div>
                  </div>
                </>
              )}

              {activeTab === 'hexo' && (
                <div className="space-y-6">
                  <div className="bg-primary/10 border border-primary/20 rounded-md p-4">
                    <h4 className="text-sm font-medium text-primary mb-2">Hexo 同步状态</h4>
                    <div className="text-sm text-foreground space-y-1">
                      <p>上次同步时间: {syncStatus?.last_sync_time ? new Date(syncStatus.last_sync_time).toLocaleString() : '从未同步'}</p>
                      <p>数据库发布文章数: {syncStatus?.total_posts ?? 0}</p>
                      <p>本地 Markdown 文件数: {syncStatus?.local_files ?? 0}</p>
                      {syncStatus?.pending_sync && (
                        <p className="text-orange-600 font-medium mt-2">提示：有待同步的变更</p>
                      )}
                    </div>
                  </div>

                  <button
                    onClick={handleSyncHexo}
                    className="px-4 py-2 bg-primary text-primary-foreground rounded hover:bg-primary/90 transition-colors flex items-center gap-2 text-sm font-medium"
                  >
                    <RefreshCw className="w-4 h-4" />
                    手动触发全量同步
                  </button>
                </div>
              )}

              {activeTab !== 'hexo' && (
                <div className="pt-6">
                  <button
                    onClick={handleSave}
                    disabled={saving}
                    className="px-4 py-2 bg-primary text-primary-foreground rounded hover:bg-primary/90 transition-colors flex items-center gap-2 text-sm font-medium disabled:opacity-50"
                  >
                    <Save className="w-4 h-4" />
                    {saving ? '保存中...' : '保存设置'}
                  </button>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
